#!/usr/bin/env python3
import sqlite3
import uuid
import json
import datetime
import os
import random

db_path = os.path.abspath(os.path.join(os.path.dirname(__file__), "../backend/autodidact.db"))
print(f"Connecting to database at {db_path}...")

conn = sqlite3.connect(db_path)
conn.execute("PRAGMA foreign_keys = ON;")
cursor = conn.cursor()

# Reset existing database records
print("Resetting tables...")
cursor.execute("DELETE FROM focus_items;")
cursor.execute("DELETE FROM step_progress;")
cursor.execute("DELETE FROM steps;")
cursor.execute("DELETE FROM modules;")
cursor.execute("DELETE FROM syllabi;")
conn.commit()

now = datetime.datetime.now(datetime.timezone.utc)

def iso_date(days_ago=0, hours_ago=0, minutes_ago=0):
    t = now - datetime.timedelta(days=days_ago, hours=hours_ago, minutes=minutes_ago)
    return t.strftime("%Y-%m-%d %H:%M:%S")

queues = [
    {
        "title": "Distributed Systems & Consensus",
        "slug": "distributed-systems-consensus",
        "description": "Core papers, fault-tolerant replication models, Raft consensus, and practical multi-node architectures.",
        "modules": [
            {
                "title": "Foundational Texts & Replicated Logs",
                "steps": [
                    ("Designing Data-Intensive Applications Ch. 5 Replication", "article", ["https://dataintensive.net/", "https://github.com/ept/ddia-references"], "Single-leader vs multi-leader vs leaderless replication. Handling failover and replication lag.", "completed", "Key takeaway: Read-after-write consistency requires routing user edits to leader or reading replica with monotonic timestamps.", 15),
                    ("Designing Data-Intensive Applications Ch. 6 Partitioning", "article", ["https://dataintensive.net/"], "Key-range vs hash partitioning. Consistent hashing and rebalancing partitions dynamically.", "completed", "Consistent hashing with virtual nodes ensures even partition distribution when nodes join or leave.", 13),
                    ("Designing Data-Intensive Applications Ch. 7 Transactions", "article", ["https://dataintensive.net/", "https://jepsen.io/analyses"], "ACID guarantees, dirty reads, non-repeatable reads, phantom reads, SSI (Serializable Snapshot Isolation).", "completed", "SSI detects write skews by tracking anti-dependency cycles in transaction conflict graphs.", 11),
                    ("Designing Data-Intensive Applications Ch. 8 Troubles with Distributed Systems", "article", ["https://dataintensive.net/"], "Network partitions, clock skews, NTP synchronization limits, Byzantine faults vs crash-recovery.", "completed", "Never rely on monotonic wall clocks for ordering across nodes without TrueTime bounds.", 9),
                    ("Designing Data-Intensive Applications Ch. 9 Consistency & Consensus", "article", ["https://dataintensive.net/"], "Linearizability vs serializability, total order broadcast, 2PC vs distributed consensus.", "in_progress", "Consensus is equivalent to total order broadcast. Currently verifying quorum requirements W + R > N.", 2),
                    ("Lamport Timers and Vector Clocks in Practice", "article", ["https://lamport.azurewebsites.net/pubs/time-clocks.pdf"], "Logical timestamps, partial vs total ordering, causal broadcast, and happens-before relations.", "completed", "Vector clocks capture exact causality: a -> b iff V(a) < V(b) component-wise.", 17),
                    ("Chandy-Lamport Distributed Snapshot Algorithm", "article", ["https://en.wikipedia.org/wiki/Chandy%E2%80%93Lamport_algorithm"], "Recording global state without stopping the system using marker messages on FIFO channels.", "completed", "Markers separate messages into those sent before the snapshot and those sent after.", 8),
                    ("Brewer's CAP Theorem & PACELC Trade-Off Model", "article", ["https://www.cs.umd.edu/~abadi/papers/abadi-pacelc.pdf"], "Partition tolerance vs availability and consistency; latency vs consistency in normal operation.", "completed", "PACELC clarifies trade-offs: if Partition, choose Availability or Consistency; Else, choose Latency or Consistency.", 21)
                ]
            },
            {
                "title": "Consensus Algorithms in Depth",
                "steps": [
                    ("In Search of an Understandable Consensus Algorithm (Raft Paper)", "article", ["https://raft.github.io/raft.pdf", "https://raft.github.io/"], "Ongaro and Ousterhout 2014 paper. Leader election, log replication, randomized timers, safety invariants.", "completed", "Leader completeness property: if an entry is committed in a given term, that entry will be present in logs of leaders for all higher terms.", 6),
                    ("Visualizing Raft Consensus Cluster Simulation", "video", ["https://thesecretlivesofdata.com/raft/"], "Interactive walkthrough of split votes, heartbeat intervals, and network partition recovery.", "completed", "Heartbeat interval (50ms) must be an order of magnitude smaller than election timeout (150-300ms).", 5),
                    ("Paxos Made Simple (Leslie Lamport)", "article", ["https://lamport.azurewebsites.net/pubs/paxos-simple.pdf"], "Phase 1a/1b prepare-promise, Phase 2a/2b accept-accepted. Distinguishing Paxos rounds from log indices.", "in_progress", "Phase 1 discovers already accepted proposals; Phase 2 commits the highest-numbered proposal.", 1),
                    ("Building a Minimal Raft Node from Scratch in Go", "project", ["https://github.com/hashicorp/raft", "https://eli.thegreenplace.net/2020/implementing-raft-part-1-the-basics/"], "Implement RPC handling, state machine transitions (Follower, Candidate, Leader), election timeouts, and commit index tracking.", "pending", "", 0),
                    ("Raft Joint Consensus & Cluster Membership Changes", "article", ["https://raft.github.io/raft.pdf"], "Handling single-server vs joint consensus configuration changes C_old,new without split-brain.", "pending", "", 0),
                    ("Viewstamped Replication Revisited", "article", ["https://pmg.csail.mit.edu/papers/vr-revisited.pdf"], "Oki and Liskov. Primary-backup replication protocol with view changes. Predecessor to Paxos/Raft.", "skipped", "Skipped initial pass to focus on Raft first; return after building prototype.", 10),
                    ("Multi-Paxos Protocol Optimization & Lease Grants", "article", ["https://en.wikipedia.org/wiki/Paxos_(computer_science)#Multi-Paxos"], "Amortizing Phase 1 rounds across multiple log entries using stable leader leases.", "pending", "", 0)
                ]
            },
            {
                "title": "Industrial Storage & Spanner Architecture",
                "steps": [
                    ("The Google File System (GFS) Architecture Paper", "article", ["https://research.google/pubs/the-google-file-system/"], "Ghemawat et al. 2003. Master node metadata, 64MB chunk sizes, record appends, relaxed consistency model.", "completed", "Chunk size of 64MB drastically reduces master memory footprint and client metadata requests.", 10),
                    ("Bigtable: A Distributed Storage System for Structured Data", "article", ["https://research.google/pubs/bigtable-a-distributed-storage-system-for-structured-data/"], "Sparse multidimensional sorted map: row key, column key, timestamp. SSTables, Memtables, Chubby locks.", "completed", "Locality groups allow segregating frequently accessed column families into separate SSTables.", 7),
                    ("Spanner: Google's Globally-Distributed Database", "article", ["https://research.google/pubs/spanner-googles-globally-distributed-database/"], "Corbett et al. 2012. TrueTime API with GPS and atomic clocks, externally consistent distributed transactions.", "in_progress", "Wait-out-uncertainty: transactions delay commit reply until 2 * epsilon time elapses to guarantee absolute linearizability.", 2),
                    ("CockroachDB Design & Multi-Raft Architecture", "article", ["https://www.cockroachlabs.com/docs/stable/architecture/overview.html"], "Range-based sharding with independent Raft consensus groups, hybrid logical clocks (HLC).", "pending", "", 0),
                    ("Amazon Dynamo: High Available Key-Value Storage", "article", ["https://www.allthingsdistributed.com/files/amazon-dynamo-sosp2007.pdf"], "Consistent hashing with virtual nodes, sloppy quorums, hinted handoff, and read repair with Merkle trees.", "completed", "Sloppy quorums allow writes to proceed on non-primary nodes during network splits; reconciled on read repair.", 14)
                ]
            }
        ]
    },
    {
        "title": "Database Internals & Storage Engines",
        "slug": "database-internals-storage-engines",
        "description": "LSM trees, B+ trees, WAL write-ahead logs, concurrency control, buffer pool managers, and SQLite internals.",
        "modules": [
            {
                "title": "Disk-Based Storage & Buffer Pool Managers",
                "steps": [
                    ("CMU 15-445: Lecture 1 Relational Model & Storage", "video", ["https://15445.courses.cs.cmu.edu/fall2024/"], "Database storage hierarchy, slotted-page architecture, tuple storage layout, and record identifiers.", "completed", "Slotted pages keep an array of header slot offsets at top and grow tuples backwards from bottom.", 18),
                    ("CMU 15-445: Lecture 2 Buffer Pool Managers & Replacement Policies", "video", ["https://15445.courses.cs.cmu.edu/fall2024/"], "Frame tables, page pinning, dirty page tracking, Clock sweep, LRU-K replacement algorithm.", "completed", "LRU-K tracks distance between current reference and K-th previous reference to prevent scan thrashing.", 16),
                    ("Building a Clock Sweep Buffer Pool in Go", "project", ["https://github.com/cmu-db/bustub"], "Implement memory frame descriptor, frame table lookup map, and dirty flag flush worker.", "completed", "Implemented with sync.RWMutex per page frame to minimize global lock contention.", 12),
                    ("Row-Oriented (NSM) vs Column-Oriented (DSM) Storage", "article", ["https://15445.courses.cs.cmu.edu/fall2024/notes/03-storage2.pdf"], "OLTP vs OLAP access patterns. Vectorized execution and compression benefits of column stores.", "completed", "Columnar stores maximize memory bandwidth by loading only queried attributes into CPU cache.", 15)
                ]
            },
            {
                "title": "Indexing: B+ Trees vs LSM Trees",
                "steps": [
                    ("B+ Tree Indexing Architecture & Split / Merge Algorithms", "article", ["https://15445.courses.cs.cmu.edu/fall2024/notes/07-bplus-trees.pdf"], "Node pointers, sibling pointers, binary search within pages, leaf node linked list.", "completed", "B+ trees keep all values in leaf pages, maximizing fanout in interior nodes (fanout > 100).", 10),
                    ("Log-Structured Merge-Trees (LSM) Original Paper", "article", ["https://www.cs.umb.edu/~poneil/lsmtree.pdf"], "O'Neil et al. 1996. Memory component C0, disk component C1, rolling merge, append-only sequential writes.", "completed", "Converts random disk writes into high-throughput sequential writes; reads incur amplification.", 8),
                    ("RocksDB Architecture: MemTable, WriteBuffer, and SSTables", "article", ["https://github.com/facebook/rocksdb/wiki/RocksDB-Basics"], "Skiplist MemTables, Bloom filter block indexing, Level vs Tiered compaction trade-offs.", "in_progress", "Bloom filters reduce point query read amplification from O(L) disk lookups to O(1) in 99% of cases.", 3),
                    ("Building an LSM Storage Engine from Scratch in Go", "project", ["https://github.com/facebook/rocksdb", "https://skyzh.github.io/mini-lsm/"], "Implement WAL log, in-memory skiplist, immutable flush to SSTable files, and basic Leveled compaction.", "pending", "", 0),
                    ("Fractal Trees & Be-Trees for Low Write Amplification", "article", ["https://www.usenix.org/legacy/event/hotstorage11/tech/final_files/Bender.pdf"], "Buffering modifications in internal tree nodes to trade write amplification for query routing.", "pending", "", 0),
                    ("Trie-Based Radix Indexing: Adaptive Radix Trees (ART)", "article", ["https://db.in.tum.de/~leis/papers/ART.pdf"], "Compact in-memory tree nodes with 4, 16, 48, and 256 child pointer layouts.", "skipped", "Saved for in-memory caching phase after disk engine project.", 7)
                ]
            },
            {
                "title": "Write-Ahead Logging (WAL) & Recovery",
                "steps": [
                    ("ARIES Recovery Algorithm: Analysis, Redo, and Undo", "article", ["https://web.stanford.edu/class/cs345d-01/rl/aries.pdf"], "Mohan et al. Physiological logging, repeating history during Redo, Compensation Log Records (CLRs).", "completed", "CLRs prevent cascading rollbacks if a crash occurs during undo phase.", 9),
                    ("SQLite WAL File Format & Checkpointing Deep Dive", "article", ["https://www.sqlite.org/wal.html", "https://www.sqlite.org/fileformat2.html#walformat"], "Wal-index shared memory (.shm), wal frames, passive vs restart vs truncate checkpoints.", "in_progress", "WAL mode allows concurrent readers to proceed uninterrupted while a single writer appends to wal.", 1),
                    ("Implementing Write-Ahead Logging with fsync Group Commits", "project", ["https://postgresql.org/docs/current/wal-internals.html"], "Batching concurrent transaction commits into single fsync calls using a commit queue channel.", "pending", "", 0),
                    ("Two-Phase Locking (2PL) vs Multi-Version Concurrency Control (MVCC)", "article", ["https://15445.courses.cs.cmu.edu/fall2024/notes/16-concurrencycontrol.pdf"], "Rigorous 2PL deadlock detection vs MVCC snapshot isolation with undo logs.", "completed", "MVCC allows readers to never block writers and writers to never block readers.", 13)
                ]
            }
        ]
    },
    {
        "title": "Computer Networking & Protocols (TCP/IP & QUIC)",
        "slug": "computer-networking-protocols",
        "description": "TCP state machine, congestion control algorithms, socket programming, HTTP/2 multiplexing, and HTTP/3 QUIC.",
        "modules": [
            {
                "title": "Transport Layer & TCP Internals",
                "steps": [
                    ("Stanford CS144: Lab 0 Networking Warmup & ByteStream", "project", ["https://cs144.github.io/", "https://github.com/cs144"], "Writing an in-memory sliding ByteStream with capacity bounds, EOF handling, and stream writes.", "completed", "ByteStream buffer management using circular ring buffer to prevent reallocations.", 22),
                    ("Stanford CS144: Lab 1 Stream Reassembler", "project", ["https://cs144.github.io/"], "Reassembling out-of-order byte substrings with overlap detection and sequence tracking.", "completed", "Merging overlapping intervals: maintaining sorted linked list of unread byte segments.", 19),
                    ("Stanford CS144: Lab 2 TCP Receiver", "project", ["https://cs144.github.io/"], "Translating 32-bit wrapping sequence numbers to 64-bit absolute seqnos. SYN and FIN state handling.", "completed", "Unwrapping seqnos requires comparing distance to last checkpoint modulo 2^32.", 16),
                    ("Stanford CS144: Lab 3 TCP Sender", "project", ["https://cs144.github.io/"], "Tracking consecutive retransmission timeouts (RTO), exponential backoff, and zero-window probing.", "completed", "Zero window probing sends 1-byte payloads to elicit window update acknowledgments from receiver.", 13),
                    ("Stanford CS144: Lab 4 The Complete TCP Connection", "project", ["https://cs144.github.io/"], "Binding TCP Sender and Receiver to Tun/Tap device for live packet exchange with real OS kernel.", "in_progress", "Currently debugging TIME_WAIT 2MSL cleanup timer after receiving active close FIN-ACK.", 2),
                    ("TCP Fast Open (TFO): Eliminating 1-RTT Handshake Delay", "article", ["https://datatracker.ietf.org/doc/html/rfc7413"], "TFO cookie negotiation during initial SYN; sending payload bytes inside SYN packets on reconnection.", "completed", "TFO cookie cryptographically authenticated by server to prevent SYN flooding amplification attacks.", 8)
                ]
            },
            {
                "title": "TCP Congestion Control Algorithms",
                "steps": [
                    ("TCP Reno & Cubic: Slow Start, Congestion Avoidance, Fast Recovery", "article", ["https://datatracker.ietf.org/doc/html/rfc5681", "https://datatracker.ietf.org/doc/html/rfc8312"], "AIMD (Additive Increase Multiplicative Decrease), window halving on loss, cubic polynomial growth.", "completed", "CUBIC uses a cubic function of elapsed time since last loss, independent of RTT, making it fair over long fat networks.", 9),
                    ("BBR: Congestion-Based Congestion Control (Google)", "article", ["https://queue.acm.org/detail.cfm?id=3022184"], "Cardwell et al. Bottleneck Bandwidth and RTT estimation, pacing gain cycles, avoiding bufferbloat.", "completed", "BBR operates at the Kleinrock optimal point where pacing rate equals max bottleneck bandwidth and queue inflight equals BDP.", 6),
                    ("Simulating TCP BBR vs Cubic under Packet Loss & Jitter", "project", ["https://github.com/google/bbr"], "Setting up netem Linux tc rules to benchmark throughput over 2% loss and 100ms delay.", "in_progress", "Cubic drops cwnd to 70% upon single packet loss; BBR ignores random loss up to probeRTT.", 1),
                    ("Explicit Congestion Notification (ECN) & CoDel Active Queue Management", "article", ["https://datatracker.ietf.org/doc/html/rfc3168"], "Marking CE codepoint in IP header before buffers drop packets; eliminating queue delays.", "pending", "", 0)
                ]
            },
            {
                "title": "Modern Transport: QUIC & HTTP/3",
                "steps": [
                    ("RFC 9000: QUIC A UDP-Based Multiplexed and Secure Transport", "article", ["https://datatracker.ietf.org/doc/html/rfc9000"], "Eliminating head-of-line blocking, connection migration using Connection IDs, TLS 1.3 0-RTT handshake.", "completed", "Independent stream flow control prevents a dropped packet on stream A from stalling stream B.", 4),
                    ("Benchmarking HTTP/2 vs HTTP/3 over High Latency Mobile Links", "project", ["https://cloudflare.com/learning/performance/http3-vs-http2/"], "Measure page load time under simulated packet drop to compare stream head-of-line blocking.", "in_progress", "Under 2% packet loss, HTTP/3 finishes multi-asset page loads 35% faster due to stream independence.", 2),
                    ("Writing a Minimal HTTP/1.1 to HTTP/2 Reverse Proxy in Go", "project", ["https://github.com/golang/net/http2"], "Frame encoding (HEADERS, DATA, SETTINGS), HPACK header compression, and stream flow control.", "pending", "", 0)
                ]
            }
        ]
    },
    {
        "title": "Go Internals & High-Performance Engineering",
        "slug": "go-internals-high-performance",
        "description": "Go runtime scheduler (GMP model), memory allocator, garbage collector write barriers, and zero-allocation profiling.",
        "modules": [
            {
                "title": "Go Runtime Scheduler & Concurrency",
                "steps": [
                    ("Analysis of the Go Scheduler: GMP Architecture", "article", ["https://go.dev/src/runtime/proc.go", "https://morsmachine.dk/go-scheduler"], "Goroutine (G), OS thread (M), Logical processor (P). Local run queues, work stealing, sysmon preemptions.", "completed", "Each P has a 256-element lock-free circular run queue; work-stealing steals half of another P's queue.", 24),
                    ("Go Cooperative vs Non-Cooperative Goroutine Preemption", "article", ["https://medium.com/a-journey-with-go/go-goroutine-and-preemption-d6bc2aa0a201"], "Preemption via stack allocation checks vs signal-based async preemption (SIGURG in Go 1.14+).", "completed", "Async preemption sends OS signals to pause tight CPU-bound loops that do not make function calls.", 20),
                    ("Sync.Pool Internals: Victim Caches & Lock-Free Rings", "article", ["https://go.dev/src/sync/pool.go"], "Per-P private object, shared local pool ring, victim cache retention between GC cycles.", "completed", "Victim cache prevents sudden memory thrashing on GC by retaining items through one collection cycle.", 17),
                    ("Channels Internals: hchan struct, sudog wait queues, and lock contention", "article", ["https://go.dev/src/runtime/chan.go"], "Circular ring buffer, recvq/sendq doubly-linked lists, direct copy from sender stack to receiver.", "completed", "Unbuffered channel transfer copies directly from sender goroutine stack to receiver goroutine stack.", 14)
                ]
            },
            {
                "title": "Memory Allocation & Garbage Collection",
                "steps": [
                    ("Go Memory Allocator: TCMalloc, Spans, Arenas, and Slabs", "article", ["https://go.dev/src/runtime/malloc.go", "https://goog-perftools.sourceforge.net/doc/tcmalloc.html"], "Tiny (<16B), Small (<32KB), Large allocations. mcache, mcentral, and mheap tiered allocation.", "completed", "Tiny allocator coalesces tiny allocations into 16-byte blocks to avoid fragmentation.", 12),
                    ("Go Garbage Collector: Tri-Color Concurrent Mark & Sweep", "article", ["https://go.dev/blog/ismmkeynote", "https://go.dev/src/runtime/mgc.go"], "White, grey, and black sets. Hybrid write barrier (Yuasa + Dijkstra) to eliminate STW pause times.", "in_progress", "Hybrid write barrier shades shaded overwritten pointers and shades new pointers to black, maintaining weak tri-color invariant.", 3),
                    ("Tuning GOGC and GOMEMLIMIT for Predictable Latency", "article", ["https://go.dev/doc/gc-guide"], "Setting GOMEMLIMIT soft memory cap to eliminate OOM kills under container burst loads.", "completed", "GOMEMLIMIT dynamically adjusts GC frequency as heap approaches memory budget.", 7),
                    ("Escape Analysis: Stack vs Heap Allocation Rules in Go Compiler", "article", ["https://medium.com/a-journey-with-go/go-mechanisms-behind-escape-analysis-8e99a236c0c5"], "Pointers escaping across stack frames, interface{} boxing conversions, and closure variable captures.", "completed", "Returning values rather than pointers allows compiler to keep structures stack-allocated.", 11)
                ]
            },
            {
                "title": "Profiling & Low-Level Optimization",
                "steps": [
                    ("pprof & execution trace profiling: CPU, Heap, Block, Mutex", "video", ["https://go.dev/blog/pprof", "https://go.dev/blog/execution-traces-2024"], "Using go tool pprof and go tool trace to identify contention, scheduler latency, and sysmon wakeups.", "completed", "Profile-guided optimization (PGO) in Go 1.20+ yields 2-7% throughput improvements by inlining hot call paths.", 8),
                    ("Zero-Allocation JSON Parsing with Byte Slices in Go", "project", ["https://github.com/valyala/fastjson", "https://github.com/bytedance/sonic"], "Build a streaming tokenizer that scans bytes directly without string allocations or interface reflection.", "in_progress", "Scanning bytes directly without allocating strings saves 90% of heap pressure on high-QPS API routes.", 1),
                    ("Go Assembly (Plan 9) & SIMD Vectorization with AVX-512", "article", ["https://go.dev/doc/asm"], "FP/SB registers, vector registers, and writing hand-optimized assembly routines.", "pending", "", 0)
                ]
            }
        ]
    },
    {
        "title": "Operating Systems & Linux Kernel Engineering",
        "slug": "operating-systems-linux-kernel",
        "description": "Process scheduling, virtual memory paging, eBPF tracing, epoll event loops, and io_uring asynchronous I/O.",
        "modules": [
            {
                "title": "Virtual Memory & Memory Management",
                "steps": [
                    ("Linux Virtual Memory: Page Tables, TLB, and Page Faults", "article", ["https://lwn.net/Kernel/Index/", "https://www.kernel.org/doc/html/latest/admin-guide/mm/index.html"], "4-level page tables (PGD, P4D, PUD, PMD, PTE), HugePages, TLB shootdowns, copy-on-write (COW).", "completed", "Page fault handler determines if address is valid in VMA; allocates physical page frame and updates PTE.", 27),
                    ("Slab, Slub, and Slob Allocators in the Linux Kernel", "article", ["https://lwn.net/Articles/229984/"], "Object caches for kernel data structures (dentry, inode, task_struct) to eliminate fragmentation.", "completed", "SLUB allocator removes per-CPU queue metadata overhead from SLAB and uses page struct fields directly.", 23),
                    ("Writing a Minimal Unix Shell with Pipes and Fork/Exec", "project", ["https://brennan.io/2015/01/16/write-a-shell-in-c/"], "Implement fork(), execvp(), waitpid(), dup2() pipe redirection, and job control signals.", "completed", "Close pipe write ends in parent before waiting, otherwise child read will block forever waiting for EOF.", 19),
                    ("Demand Paging and Page Replacement (Clock / Second-Chance Algorithm)", "article", ["https://pages.cs.wisc.edu/~remzi/OSTEP/vm-beyondphys.pdf"], "Active and inactive page lists in Linux, kswapd background flusher, and OOM killer invocation.", "completed", "Linux splits pages into active and inactive LRU lists to protect working sets from sequential scans.", 15)
                ]
            },
            {
                "title": "High-Performance Linux I/O: epoll & io_uring",
                "steps": [
                    ("The Architecture of epoll: Edge-Triggered vs Level-Triggered", "article", ["https://man7.org/linux/man-pages/man7/epoll.7.html"], "Red-black tree file registration, ready-list linked list, EPOLLET non-blocking drain loops.", "completed", "Edge-triggered mode requires draining file descriptor until EAGAIN/EWOULDBLOCK to avoid missed events.", 14),
                    ("Efficient IO with io_uring: Jens Axboe Architecture Paper", "article", ["https://kernel.dk/io_uring.pdf"], "Submission Queue (SQ) and Completion Queue (CQ) ring buffers in shared memory, zero-syscall I/O.", "completed", "Shared memory ring buffers allow application and kernel to submit and complete requests without context switching.", 11),
                    ("Building a High-Concurrency HTTP Server with io_uring in C", "project", ["https://github.com/axboe/liburing"], "Implement IORING_OP_ACCEPT, IORING_OP_READ, and IORING_OP_WRITE with SQPOLL kernel thread.", "in_progress", "SQPOLL mode keeps kernel polling thread alive; app submits operations without making any syscalls.", 2),
                    ("Zero-Copy Networking: sendfile, splice, and MSG_ZEROCOPY", "article", ["https://lwn.net/Articles/726917/"], "Bypassing user-space memory buffers; DMA transferring page cache directly to network adapter ring.", "pending", "", 0)
                ]
            },
            {
                "title": "Linux Observability & eBPF",
                "steps": [
                    ("What is eBPF? In-Kernel Sandboxed Programmability", "article", ["https://ebpf.io/", "https://cilium.io/"], "eBPF verifier safety guarantees, JIT compiler, BPF maps, kprobes, tracepoints, and XDP.", "completed", "Verifier ensures eBPF programs terminate and never access uninitialized or out-of-bounds memory.", 7),
                    ("Writing an eBPF Tracepoint Tool to Monitor Block Device I/O Latency", "project", ["https://github.com/iovisor/bcc", "https://github.com/cilium/ebpf"], "Attach eBPF probe to block_rq_issue and block_rq_complete to compute disk latency histogram.", "in_progress", "Attached kprobe to blk_account_io_done; dumping latency histogram in 10ms logarithmic buckets.", 1),
                    ("XDP (eXpress Data Path): Dropping 10 Million Packets per Second", "article", ["https://www.iovisor.org/technology/xdp"], "Packet filtering directly inside the network driver NIC ring before sk_buff allocation.", "pending", "", 0)
                ]
            }
        ]
    },
    {
        "title": "Kubernetes & Production Platform Engineering",
        "slug": "kubernetes-platform-engineering",
        "description": "Kubelet control loop, CNI networking, CoreDNS, Custom Resource Definitions (CRDs), and Operator pattern.",
        "modules": [
            {
                "title": "Kubernetes Control Plane Architecture",
                "steps": [
                    ("Kubernetes API Server Internals: Etcd Watch & Informers", "article", ["https://kubernetes.io/docs/concepts/overview/kubernetes-api/"], "ResourceVersion pagination, Etcd MVCC watches, SharedIndexInformer delta FIFO queues.", "completed", "SharedIndexInformer reduces API server load by maintaining local cache indexed by namespace/labels.", 26),
                    ("Kube-Scheduler: Scoring, Filtering, and Prioritizing Pod Placement", "article", ["https://kubernetes.io/docs/concepts/scheduling-eviction/kube-scheduler/"], "Preemption, topology spread constraints, node affinity, taints and tolerations.", "completed", "Scheduling cycles run single-threaded per pod; binding cycles run concurrently in background.", 20),
                    ("Writing a Custom Kubernetes Controller in Go with controller-runtime", "project", ["https://book.kubebuilder.io/"], "Implement Reconcile() loop, CRD schema definition, finalizers, and owner references for garbage collection.", "completed", "Reconcile loop must be idempotent and driven solely by current cluster state vs desired state.", 16),
                    ("Kubelet Architecture: PLEG (Pod Lifecycle Event Generator) & CRI", "article", ["https://kubernetes.io/docs/concepts/architecture/nodes/"], "Container Runtime Interface (CRI), gRPC calls to containerd, cgroups v2 resource limits.", "completed", "PLEG relists container states every second via CRI and publishes event channels to kubelet sync worker.", 12)
                ]
            },
            {
                "title": "Container Networking & Service Meshes",
                "steps": [
                    ("Linux Network Namespaces, Veth Pairs, and Bridge Routing", "article", ["https://developers.redhat.com/blog/2018/10/22/introduction-to-linux-interfaces-for-virtual-networking"], "How containers get IP addresses. Veth pairs, iptables MASQUERADE, and route tables.", "completed", "Veth pair acts as a bi-directional virtual wire connecting host network namespace with pod namespace.", 11),
                    ("Cilium CNI: eBPF-Based High-Performance Pod Networking", "article", ["https://cilium.io/doc/"], "Replacing iptables and IPVS kube-proxy routing with direct eBPF socket layer forwarding.", "in_progress", "eBPF attaches directly to tc (traffic control) hook, bypassing Linux TCP/IP stack overhead for pod-to-pod traffic.", 3),
                    ("Configuring Envoy Proxy Sidecars & mTLS Handshakes", "project", ["https://www.envoyproxy.io/docs/"], "Setup dynamic cluster configuration, certificate rotation, and circuit breaking thresholds.", "pending", "", 0),
                    ("CoreDNS Architecture: Plugin Chain, Cache Prefetching, and AutoPath", "article", ["https://coredns.io/manual/toc/"], "Handling DNS query fanout in high-density clusters; avoiding search domain query amplification.", "completed", "AutoPath plugin intercepts multi-suffix search path queries on server side to reduce client UDP round-trips.", 9)
                ]
            }
        ]
    },
    {
        "title": "Rust for High-Performance Systems",
        "slug": "rust-high-performance-systems",
        "description": "Borrow checker semantics, zero-cost abstractions, Pin and Unpin async runtime internals, and SIMD optimizations.",
        "modules": [
            {
                "title": "Ownership, Lifetimes, and Memory Safety",
                "steps": [
                    ("The Rust Programming Language Ch. 4 Understanding Ownership", "article", ["https://doc.rust-lang.org/book/ch04-00-understanding-ownership.html"], "Stack vs heap, move semantics, clone vs copy, reference borrowing rules, mutable references.", "completed", "At any given time, you can have either one mutable reference or any number of immutable references.", 30),
                    ("The Rust Programming Language Ch. 10 Lifetimes & Generics", "article", ["https://doc.rust-lang.org/book/ch10-00-generics.html"], "Lifetime annotations 'a, lifetime elision rules in compiler, static lifetimes.", "completed", "Lifetime annotations tell the compiler how the generic lifetime parameters of multiple references relate to each other.", 26),
                    ("Writing a High-Throughput LRU Cache in Safe Rust", "project", ["https://github.com/moka-rs/moka"], "Implement doubly-linked list with generational arena indexing to avoid raw pointer unsafety.", "completed", "Using vector-backed slot map index removes raw pointer dereferences while maintaining O(1) operations.", 22),
                    ("Interior Mutability: Cell, RefCell, and UnsafeCell", "article", ["https://doc.rust-lang.org/std/cell/"], "Bypassing compiler immutability checks at runtime; dynamically tracked borrow flags in RefCell.", "completed", "RefCell tracks borrow count via integer flag; panics at runtime if alias borrowing rules are violated.", 17)
                ]
            },
            {
                "title": "Async Rust & Tokio Runtime Internals",
                "steps": [
                    ("Understanding Future, Poll, and Context in Rust", "article", ["https://rust-lang.github.io/async-book/"], "How state machines are generated by async/await. Wakers, Reactor pattern, and Poll::Ready vs Poll::Pending.", "completed", "A Future does nothing until polled; the Waker notifies the executor when IO readiness occurs.", 16),
                    ("Pin, Unpin, and Self-Referential Structs Explained", "article", ["https://fasterthanli.me/articles/pin-and-suffering"], "Why self-referential generator structs cannot be moved in memory once pinned. Pin<&mut T> guarantees.", "completed", "Pinning guarantees that the memory address of the struct will not change before drop is called.", 13),
                    ("Building a Mini Async TCP Server with Tokio", "project", ["https://tokio.rs/tokio/tutorial"], "Implement connection multiplexer, framed codecs, and graceful shutdown signal handling.", "in_progress", "Using CancellationToken to broadcast shutdown events across spawned tokio tasks.", 2),
                    ("Lock-Free Data Structures in Rust: Crossbeam & AtomicPtr", "article", ["https://docs.rs/crossbeam/latest/crossbeam/"], "Epoch-based memory reclamation (EBR), atomic compare-and-swap (CAS), memory orderings (Acquire/Release).", "pending", "", 0)
                ]
            }
        ]
    },
    {
        "title": "Machine Learning Systems & Inference Engines",
        "slug": "ml-systems-inference-engines",
        "description": "CUDA memory hierarchy, KV-cache management, PagedAttention, quantization (AWQ, GPTQ), and TensorRT-LLM.",
        "modules": [
            {
                "title": "Transformer Architecture & Attention Mechanisms",
                "steps": [
                    ("Attention Is All You Need (Vaswani et al. 2017)", "article", ["https://arxiv.org/abs/1706.03762"], "Scaled dot-product attention, multi-head attention, positional encodings, feed-forward layers.", "completed", "Softmax scaling factor 1/sqrt(d_k) prevents dot-product values from growing too large into vanishing gradient regions.", 28),
                    ("FlashAttention: Fast and Memory-Efficient Exact Attention with IO-Awareness", "article", ["https://arxiv.org/abs/2205.14135"], "Dao et al. Tiling attention matrix across GPU SRAM to eliminate HBM read/write bottlenecks.", "completed", "Online softmax computes normalization dynamically across blocks without materializing NxN attention matrix in HBM.", 23),
                    ("vLLM: Efficient Memory Management with PagedAttention", "article", ["https://arxiv.org/abs/2309.06180"], "PagedAttention treats KV-cache like virtual memory pages, reducing memory waste from 70% to under 4%.", "in_progress", "Continuous batching and virtual block tables allow sharing KV-cache blocks across parallel decode beams.", 1),
                    ("Rotary Position Embeddings (RoPE) Math & Implementation", "article", ["https://arxiv.org/abs/2104.09864"], "Encoding absolute position with rotation matrix in complex space; naturally decaying attention over relative distance.", "completed", "RoPE applies complex number rotation to query/key vector pairs, preserving relative position invariance.", 19)
                ]
            },
            {
                "title": "Model Quantization & Inference Runtime",
                "steps": [
                    ("AWQ: Activation-aware Weight Quantization for LLM Inference", "article", ["https://arxiv.org/abs/2306.00978"], "Protecting 1% salient weights based on activation magnitudes to maintain FP16 accuracy at 4-bit weights.", "completed", "Salient weight channels are scaled rather than quantized, preserving attention sensitivity.", 15),
                    ("Writing a Minimal GGUF Parser and Tensor Loader in C/Go", "project", ["https://github.com/ggerganov/llama.cpp"], "Parse header metadata, key-value pairs, tensor offsets, and mmap tensor byte arrays directly into memory.", "pending", "", 0),
                    ("Speculative Decoding: Small Draft Model Accelerated Verification", "article", ["https://arxiv.org/abs/2211.17192"], "Using small 1B draft model to guess K tokens, verified in single forward pass by large 70B target model.", "pending", "", 0)
                ]
            }
        ]
    },
    {
        "title": "Compilers, Bytecode VMs & JIT Execution",
        "slug": "compilers-bytecode-vms-jit",
        "description": "Lexing, recursive descent parsing, AST representations, SSA intermediate representation, and register allocation.",
        "modules": [
            {
                "title": "Scanning, Parsing, and AST Representation",
                "steps": [
                    ("Crafting Interpreters Part 1: Tree-Walk Interpreter (jlox)", "article", ["https://craftinginterpreters.com/"], "Scanning tokens, CFG grammars, recursive descent parsing, evaluating expressions, lexical scoping environment.", "completed", "Lexical scopes modeled as an environment chain of parent hash maps.", 32),
                    ("Pratt Parsing: Top-Down Operator Precedence Algorithms", "article", ["https://matklad.github.io/2020/04/13/simple-but-powerful-pratt-parsing.html"], "Infix vs prefix binding powers to eliminate complex grammar rules for operator precedence.", "completed", "Binding power table simplifies parsing binary, prefix, and postfix operators in O(N).", 29),
                    ("Crafting Interpreters Part 2: Bytecode Virtual Machine (clox)", "article", ["https://craftinginterpreters.com/a-bytecode-virtual-machine.html"], "Chunk bytecode array, opcode dispatch loop, value stack, constant pool, and NaN-boxing.", "completed", "NaN-boxing packs doubles, booleans, nil, and heap pointers into 64-bit IEEE 754 representations.", 24),
                    ("Building a Stack-Based Bytecode VM in Go", "project", ["https://compilerbook.com/"], "Implement compiler bytecode emitter, instruction decoder, and execution stack for arithmetic and variables.", "completed", "Threaded code dispatch using function pointer tables reduces branch misprediction.", 20)
                ]
            },
            {
                "title": "Intermediate Representations & Optimizations",
                "steps": [
                    ("Static Single Assignment (SSA) Form & Dominance Frontiers", "article", ["https://www.cs.cmu.edu/~fp/courses/15411-f14/lectures/06-ssa.pdf"], "Dominator trees, phi-functions, Cytron dominance frontier algorithm for minimal SSA placement.", "completed", "Phi functions resolve merged variable definitions across incoming control flow branches.", 17),
                    ("Common Subexpression Elimination (CSE) & Dead Code Elimination", "article", ["https://www.cs.cornell.edu/courses/cs6120/2020fa/self-directed/dce/"], "Available expressions analysis, dataflow framework, worklist algorithm for dead instruction removal.", "completed", "Global value numbering (GVN) identifies congruent expressions across entire basic blocks.", 13),
                    ("Register Allocation: Chaitin-Briggs Graph Coloring", "article", ["https://www.cs.utexas.edu/~pingali/CS380C/2010/lectures/regalloc.pdf"], "Interference graphs, Kempe k-coloring heuristic, spilling heuristics based on loop nesting depth.", "in_progress", "Variables with disjoint live ranges share the same physical CPU register without spill overhead.", 2),
                    ("Writing an LLVM IR Frontend for a Tiny Functional Language", "project", ["https://llvm.org/docs/tutorial/MyFirstLanguageFrontend/index.html"], "Generate LLVM IR using llvm-cgo bindings: basic blocks, phi nodes, and JIT compilation to machine code.", "pending", "", 0)
                ]
            }
        ]
    },
    {
        "title": "Modern Web Performance & Browser Architecture",
        "slug": "web-performance-browser-architecture",
        "description": "Blink rendering pipeline, CSS layout geometry, V8 Turbofan optimization pipeline, and WebAssembly execution.",
        "modules": [
            {
                "title": "Browser Rendering Engine Internals",
                "steps": [
                    ("Inside Look at Modern Web Browser (Part 1 to 4 by Mariko Kosaka)", "article", ["https://developer.chrome.com/blog/inside-browser-part1/"], "Browser process, renderer process, GPU process, compositor thread, and rasterization pipeline.", "completed", "Compositor thread runs independently of JS main thread to ensure smooth 60fps scrolling.", 25),
                    ("Layout, Paint, and Composite: Avoiding Layout Thrashing", "article", ["https://web.dev/rendering-performance/"], "Forced synchronous layouts, style recalculation cascades, transform/opacity compositor layers.", "completed", "Reading layout properties (offsetWidth) right after modifying styles forces synchronous style recalculation.", 21),
                    ("V8 Engine: Ignition Interpreter and TurboFan JIT Compiler", "article", ["https://v8.dev/docs/turbofan"], "Hidden classes (Shapes), Inline Caches (IC), polymorphic vs monomorphic call site deoptimizations.", "completed", "Monomorphic call sites allow V8 to emit direct machine code instructions bypassing property hash lookups.", 16),
                    ("Benchmarking Memory Allocation and Cache Lines in JS Typed Arrays", "project", ["https://developer.mozilla.org/en-US/docs/Web/JavaScript/Typed_arrays"], "Comparing ArrayBuffer contiguous memory access with regular V8 pointer array objects under heavy load.", "in_progress", "Float64Array achieves 8x faster traversal due to CPU L1 cache line prefetching.", 1)
                ]
            },
            {
                "title": "WebAssembly & Edge Runtime Engineering",
                "steps": [
                    ("WebAssembly Core Specification & Stack Machine Execution Model", "article", ["https://webassembly.github.io/spec/core/"], "Wasm linear memory, table indirection, structured control flow blocks, and instruction validation.", "completed", "Wasm control flow uses structured blocks rather than arbitrary gotos, making validation single-pass O(N).", 11),
                    ("Compiling Rust to Wasm with wasm-pack and wasm-bindgen", "project", ["https://rustwasm.github.io/docs/book/"], "Exporting memory buffers, calling host JS imports, and benchmarks against pure JS implementations.", "pending", "", 0)
                ]
            }
        ]
    },
    {
        "title": "System Design & High-Scale Real-Time Architecture",
        "slug": "system-design-realtime-architecture",
        "description": "Real-time event streams, Kafka partitions, WebSocket gateways, geo-replicated caches, and idempotency keys.",
        "modules": [
            {
                "title": "Event-Driven Messaging & Stream Processing",
                "steps": [
                    ("Apache Kafka Architecture: Log Segments, Consumer Groups, and Offsets", "article", ["https://kafka.apache.org/documentation/#design"], "Zero-copy sendfile syscall, sequential disk writes, partition rebalancing, and exactly-once semantics.", "completed", "Kafka avoids user-space memory copies by using Linux sendfile to stream disk page cache directly to network socket.", 28),
                    ("Building an Idempotent Payment Webhook Consumer in Go", "project", ["https://stripe.com/docs/webhooks/best-practices"], "Transactional outbox pattern, distributed locks, database unique constraints, and exponential backoff retry.", "completed", "Transactional outbox guarantees atomicity between database business mutations and message queue publications.", 24),
                    ("Redis Clustering: Hash Slots, Gossip Protocol, and Failover", "article", ["https://redis.io/docs/management/scaling/"], "16384 hash slots, CRC16 hashing, MOVED and ASK redirects, asynchronous replication risk.", "completed", "Cluster nodes exchange heartbeat ping/pong packets every second using a binary gossip protocol.", 18),
                    ("Designing a Globally Distributed Real-Time Collaborative Canvas", "article", ["https://figma.com/blog/how-figmas-multiplayer-technology-works/"], "CRDTs (Conflict-free Replicated Data Types) vs Operational Transformation (OT) over WebSockets.", "in_progress", "State-based vs operation-based CRDTs: evaluating Yjs and Automerge vector clock convergence under partition.", 3),
                    ("Building a High-Performance WebSocket Gateway with epoll in Go", "project", ["https://github.com/gobwas/ws"], "Handling 100k persistent concurrent connections with zero-allocation frame decoders.", "pending", "", 0)
                ]
            }
        ]
    },
    {
        "title": "Applied Cryptography & Security Protocols",
        "slug": "applied-cryptography-security-protocols",
        "description": "Symmetric ciphers, authenticated encryption (AEAD), elliptic curve Diffie-Hellman, and TLS 1.3 handshakes.",
        "modules": [
            {
                "title": "Symmetric Ciphers & Message Authentication",
                "steps": [
                    ("AES-GCM Authenticated Encryption: Block Cipher Modes and GHASH", "article", ["https://csrc.nist.gov/publications/detail/sp/800-38d/final"], "Galois/Counter Mode (GCM), counter increment, GHASH polynomial evaluation, avoiding nonce reuse.", "completed", "Reusing a nonce in AES-GCM destroys authentication integrity and allows recovering the authentication subkey.", 31),
                    ("ChaCha20-Poly1305: Bernstein Stream Cipher and MAC", "article", ["https://datatracker.ietf.org/doc/html/rfc8439"], "Quarter round permutations, state matrix rotation, Poly1305 clamp masks, constant-time execution.", "completed", "ChaCha20 is designed to execute in constant time on all hardware without relying on CPU lookup tables.", 27),
                    ("Cryptopals Crypto Challenges: Sets 1 to 3", "project", ["https://cryptopals.com/"], "Breaking repeating-key XOR, CBC bit-flipping attack, and CBC padding oracle decryption exploit.", "completed", "Padding oracle exploits side-channel differences in decryption errors to decrypt ciphertext without knowing the key.", 23)
                ]
            },
            {
                "title": "Asymmetric Cryptography & Key Exchange",
                "steps": [
                    ("Elliptic Curve Diffie-Hellman (ECDH) on Curve25519 (X25519)", "article", ["https://cr.yp.to/ecdh.html"], "Montgomery curve arithmetic, scalar multiplication, Montgomery ladder constant-time algorithm.", "completed", "Montgomery ladder computes scalar multiplication with fixed sequence of field operations regardless of bit values.", 15),
                    ("The TLS 1.3 Handshake Protocol (RFC 8446)", "article", ["https://datatracker.ietf.org/doc/html/rfc8446"], "1-RTT full handshake, 0-RTT early data replay risks, removal of static RSA and vulnerable ciphers.", "in_progress", "Server responds with server_hello and encrypted extensions in single round trip, encrypting all handshake metadata.", 2),
                    ("Writing a Toy TLS 1.3 Client Handshake in Go from Raw TCP Sockets", "project", ["https://github.com/google/fscrypt"], "Construct ClientHello, parse ServerHello, compute shared secret with X25519, derive handshake keys with HKDF.", "pending", "", 0)
                ]
            }
        ]
    }
]

total_steps = 0
focus_candidates = []

for q_idx, q in enumerate(queues):
    syl_id = str(uuid.uuid4())
    s_created = iso_date(days_ago=random.randint(30, 60))
    cursor.execute("""
        INSERT INTO syllabi (id, title, slug, description, is_template, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    """, (syl_id, q["title"], q["slug"], q["description"], 0, s_created, s_created))

    for m_idx, mod in enumerate(q["modules"]):
        mod_id = str(uuid.uuid4())
        m_created = iso_date(days_ago=random.randint(15, 30))
        cursor.execute("""
            INSERT INTO modules (id, syllabus_id, title, description, order_index, created_at)
            VALUES (?, ?, ?, ?, ?, ?)
        """, (mod_id, syl_id, mod["title"], "", m_idx, m_created))

        for step_idx, step_tuple in enumerate(mod["steps"]):
            total_steps += 1
            title, st_type, resources, content, status, notes, days_ago = step_tuple
            step_id = str(uuid.uuid4())
            step_created = iso_date(days_ago=max(days_ago + 1, 1))
            res_json = json.dumps(resources)
            primary_url = resources[0] if resources else ""

            cursor.execute("""
                INSERT INTO steps (id, module_id, title, type, content, resource_url, resources, acceptance_criteria, estimated_minutes, order_index, created_at)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            """, (
                step_id, mod_id, title, st_type, content,
                primary_url, res_json, "", random.choice([30, 45, 60, 90]),
                step_idx, step_created
            ))

            completed_at = iso_date(days_ago=days_ago) if status == "completed" else None
            cursor.execute("""
                INSERT INTO step_progress (id, step_id, status, notes, submission_url, completed_at, created_at)
                VALUES (?, ?, ?, ?, ?, ?, ?)
            """, (
                str(uuid.uuid4()), step_id, status, notes, "", completed_at, step_created
            ))

            if status in ("in_progress", "completed", "pending"):
                focus_candidates.append({
                    "step_id": step_id,
                    "title": title,
                    "resources": resources,
                    "status": "in_progress" if status == "in_progress" else ("completed" if status == "completed" else "todo"),
                    "notes": notes,
                    "completed_at": completed_at,
                    "days_ago": days_ago
                })

print(f"Total inserted: {len(queues)} syllabi, {total_steps} steps.")

# Seed focus items
in_prog = [c for c in focus_candidates if c["status"] == "in_progress"]
completed = [c for c in focus_candidates if c["status"] == "completed"]
todo = [c for c in focus_candidates if c["status"] == "todo"]

selected_focus = []
# 3 Doing
selected_focus.extend(in_prog[:3])
# 5 To Do
selected_focus.extend(todo[:5])
# 6 Finished Log
selected_focus.extend(completed[:6])

for idx, f_item in enumerate(selected_focus):
    focus_id = str(uuid.uuid4())
    res_json = json.dumps(f_item["resources"])
    primary_url = f_item["resources"][0] if f_item["resources"] else ""
    created = iso_date(days_ago=max(f_item["days_ago"], 1))
    updated = iso_date(days_ago=max(f_item["days_ago"] - 1, 0))

    cursor.execute("""
        INSERT INTO focus_items (id, step_id, custom_title, resource_url, resources, period, status, notes, order_index, completed_at, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    """, (
        focus_id,
        f_item["step_id"],
        "",
        primary_url,
        res_json,
        "next",
        f_item["status"],
        f_item["notes"],
        idx,
        f_item["completed_at"],
        created,
        updated
    ))

conn.commit()
conn.close()
print(f"Successfully generated database with {total_steps} items and {len(selected_focus)} What Next focus items!")
