package slurm

// Reasons for job to be pending
const (
	WAIT_NO_REASON                      uint32 = iota // not set or job not pending
	WAIT_PRIORITY                                     // higher priority jobs exist
	WAIT_DEPENDENCY                                   // dependent job has not completed
	WAIT_RESOURCES                                    // required resources not available
	WAIT_PART_NODE_LIMIT                              // request exceeds partition node limit
	WAIT_PART_TIME_LIMIT                              // request exceeds partition time limit
	WAIT_PART_DOWN                                    // requested partition is down
	WAIT_PART_INACTIVE                                // requested partition is inactive
	WAIT_HELD                                         // job is held by administrator
	WAIT_TIME                                         // job waiting for specific begin time
	WAIT_LICENSES                                     // job is waiting for licenses
	WAIT_ASSOC_JOB_LIMIT                              // user/account job limit reached
	WAIT_ASSOC_RESOURCE_LIMIT                         // user/account resource limit reached
	WAIT_ASSOC_TIME_LIMIT                             // user/account time limit reached
	WAIT_RESERVATION                                  // reservation not available
	WAIT_NODE_NOT_AVAIL                               // required node is DOWN or DRAINED
	WAIT_HELD_USER                                    // job is held by user
	DEFUNCT_WAIT_17                                   // free for reuse
	FAIL_DEFER                                        // individual submit time sched deferred
	FAIL_DOWN_PARTITION                               // partition for job is DOWN
	FAIL_DOWN_NODE                                    // some node in the allocation failed
	FAIL_BAD_CONSTRAINTS                              // constraints can not be satisfied
	FAIL_SYSTEM                                       // slurm system failure
	FAIL_LAUNCH                                       // unable to launch job
	FAIL_EXIT_CODE                                    // exit code was non-zero
	FAIL_TIMEOUT                                      // reached end of time limit
	FAIL_INACTIVE_LIMIT                               // reached slurm InactiveLimit
	FAIL_ACCOUNT                                      // invalid account
	FAIL_QOS                                          // invalid QOS
	WAIT_QOS_THRES                                    // required QOS threshold has been breached
	WAIT_QOS_JOB_LIMIT                                // QOS job limit reached
	WAIT_QOS_RESOURCE_LIMIT                           // QOS resource limit reached
	WAIT_QOS_TIME_LIMIT                               // QOS time limit reached
	FAIL_SIGNAL                                       // raised a signal that caused it to exit
	DEFUNCT_WAIT_34                                   // free for reuse
	WAIT_CLEANING                                     // If a job is requeued and it isstill cleaning up from the last run.
	WAIT_PROLOG                                       // Prolog is running
	WAIT_QOS                                          // QOS not allowed
	WAIT_ACCOUNT                                      // Account not allowed
	WAIT_DEP_INVALID                                  // Dependency condition invalid or neversatisfied
	WAIT_QOS_GRP_CPU                                  // QOS GrpTRES exceeded (CPU)
	WAIT_QOS_GRP_CPU_MIN                              // QOS GrpTRESMins exceeded (CPU)
	WAIT_QOS_GRP_CPU_RUN_MIN                          // QOS GrpTRESRunMins exceeded (CPU)
	WAIT_QOS_GRP_JOB                                  // QOS GrpJobs exceeded
	WAIT_QOS_GRP_MEM                                  // QOS GrpTRES exceeded (Memory)
	WAIT_QOS_GRP_NODE                                 // QOS GrpTRES exceeded (Node)
	WAIT_QOS_GRP_SUB_JOB                              // QOS GrpSubmitJobs exceeded
	WAIT_QOS_GRP_WALL                                 // QOS GrpWall exceeded
	WAIT_QOS_MAX_CPU_PER_JOB                          // QOS MaxTRESPerJob exceeded (CPU)
	WAIT_QOS_MAX_CPU_MINS_PER_JOB                     // QOS MaxTRESMinsPerJob exceeded (CPU)
	WAIT_QOS_MAX_NODE_PER_JOB                         // QOS MaxTRESPerJob exceeded (Node)
	WAIT_QOS_MAX_WALL_PER_JOB                         // QOS MaxWallDurationPerJob exceeded
	WAIT_QOS_MAX_CPU_PER_USER                         // QOS MaxTRESPerUser exceeded (CPU)
	WAIT_QOS_MAX_JOB_PER_USER                         // QOS MaxJobsPerUser exceeded
	WAIT_QOS_MAX_NODE_PER_USER                        // QOS MaxTRESPerUser exceeded (Node)
	WAIT_QOS_MAX_SUB_JOB                              // QOS MaxSubmitJobsPerUser exceeded
	WAIT_QOS_MIN_CPU                                  // QOS MinTRESPerJob not reached (CPU)
	WAIT_ASSOC_GRP_CPU                                // ASSOC GrpTRES exceeded (CPU)
	WAIT_ASSOC_GRP_CPU_MIN                            // ASSOC GrpTRESMins exceeded (CPU)
	WAIT_ASSOC_GRP_CPU_RUN_MIN                        // ASSOC GrpTRESRunMins exceeded (CPU)
	WAIT_ASSOC_GRP_JOB                                // ASSOC GrpJobs exceeded
	WAIT_ASSOC_GRP_MEM                                // ASSOC GrpTRES exceeded (Memory)
	WAIT_ASSOC_GRP_NODE                               // ASSOC GrpTRES exceeded (Node)
	WAIT_ASSOC_GRP_SUB_JOB                            // ASSOC GrpSubmitJobs exceeded
	WAIT_ASSOC_GRP_WALL                               // ASSOC GrpWall exceeded
	WAIT_ASSOC_MAX_JOBS                               // ASSOC MaxJobs exceeded
	WAIT_ASSOC_MAX_CPU_PER_JOB                        // ASSOC MaxTRESPerJob exceeded (CPU)
	WAIT_ASSOC_MAX_CPU_MINS_PER_JOB                   // ASSOC MaxTRESMinsPerJobexceeded (CPU)
	WAIT_ASSOC_MAX_NODE_PER_JOB                       // ASSOC MaxTRESPerJob exceeded (NODE)
	WAIT_ASSOC_MAX_WALL_PER_JOB                       // ASSOC MaxWallDurationPerJobexceeded
	WAIT_ASSOC_MAX_SUB_JOB                            // ASSOC MaxSubmitJobsPerUser exceeded
	WAIT_MAX_REQUEUE                                  // MAX_BATCH_REQUEUE reached
	WAIT_ARRAY_TASK_LIMIT                             // job array running task limit
	WAIT_BURST_BUFFER_RESOURCE                        // Burst buffer resources
	WAIT_BURST_BUFFER_STAGING                         // Burst buffer file stage-in
	FAIL_BURST_BUFFER_OP                              // Burst buffer operation failure
	DEFUNCT_WAIT_76                                   // free for reuse
	DEFUNCT_WAIT_77                                   // free for reuse
	WAIT_ASSOC_GRP_UNK                                // ASSOC GrpTRES exceeded(Unknown)
	WAIT_ASSOC_GRP_UNK_MIN                            // ASSOC GrpTRESMins exceeded(Unknown)
	WAIT_ASSOC_GRP_UNK_RUN_MIN                        // ASSOC GrpTRESRunMins exceeded(Unknown)
	WAIT_ASSOC_MAX_UNK_PER_JOB                        // ASSOC MaxTRESPerJob exceeded(Unknown)
	WAIT_ASSOC_MAX_UNK_PER_NODE                       // ASSOC MaxTRESPerNode exceeded(Unknown)
	WAIT_ASSOC_MAX_UNK_MINS_PER_JOB                   // ASSOC MaxTRESMinsPerJobexceeded (Unknown)
	WAIT_ASSOC_MAX_CPU_PER_NODE                       // ASSOC MaxTRESPerNode exceeded (CPU)
	WAIT_ASSOC_GRP_MEM_MIN                            // ASSOC GrpTRESMins exceeded(Memory)
	WAIT_ASSOC_GRP_MEM_RUN_MIN                        // ASSOC GrpTRESRunMins exceeded(Memory)
	WAIT_ASSOC_MAX_MEM_PER_JOB                        // ASSOC MaxTRESPerJob exceeded (Memory)
	WAIT_ASSOC_MAX_MEM_PER_NODE                       // ASSOC MaxTRESPerNode exceeded (CPU)
	WAIT_ASSOC_MAX_MEM_MINS_PER_JOB                   // ASSOC MaxTRESMinsPerJobexceeded (Memory)
	WAIT_ASSOC_GRP_NODE_MIN                           // ASSOC GrpTRESMins exceeded (Node)
	WAIT_ASSOC_GRP_NODE_RUN_MIN                       // ASSOC GrpTRESRunMins exceeded (Node)
	WAIT_ASSOC_MAX_NODE_MINS_PER_JOB                  // ASSOC MaxTRESMinsPerJobexceeded (Node)
	WAIT_ASSOC_GRP_ENERGY                             // ASSOC GrpTRES exceeded(Energy)
	WAIT_ASSOC_GRP_ENERGY_MIN                         // ASSOC GrpTRESMins exceeded(Energy)
	WAIT_ASSOC_GRP_ENERGY_RUN_MIN                     // ASSOC GrpTRESRunMins exceeded(Energy)
	WAIT_ASSOC_MAX_ENERGY_PER_JOB                     // ASSOC MaxTRESPerJob exceeded(Energy)
	WAIT_ASSOC_MAX_ENERGY_PER_NODE                    // ASSOC MaxTRESPerNodeexceeded (Energy)
	WAIT_ASSOC_MAX_ENERGY_MINS_PER_JOB                // ASSOC MaxTRESMinsPerJobexceeded (Energy)
	WAIT_ASSOC_GRP_GRES                               // ASSOC GrpTRES exceeded (GRES)
	WAIT_ASSOC_GRP_GRES_MIN                           // ASSOC GrpTRESMins exceeded (GRES)
	WAIT_ASSOC_GRP_GRES_RUN_MIN                       // ASSOC GrpTRESRunMins exceeded (GRES)
	WAIT_ASSOC_MAX_GRES_PER_JOB                       // ASSOC MaxTRESPerJob exceeded (GRES)
	WAIT_ASSOC_MAX_GRES_PER_NODE                      // ASSOC MaxTRESPerNode exceeded (GRES)
	WAIT_ASSOC_MAX_GRES_MINS_PER_JOB                  // ASSOC MaxTRESMinsPerJobexceeded (GRES)
	WAIT_ASSOC_GRP_LIC                                // ASSOC GrpTRES exceeded(license)
	WAIT_ASSOC_GRP_LIC_MIN                            // ASSOC GrpTRESMins exceeded(license)
	WAIT_ASSOC_GRP_LIC_RUN_MIN                        // ASSOC GrpTRESRunMins exceeded(license)
	WAIT_ASSOC_MAX_LIC_PER_JOB                        // ASSOC MaxTRESPerJob exceeded(license)
	WAIT_ASSOC_MAX_LIC_MINS_PER_JOB                   // ASSOC MaxTRESMinsPerJob exceeded(license)
	WAIT_ASSOC_GRP_BB                                 // ASSOC GrpTRES exceeded(burst buffer)
	WAIT_ASSOC_GRP_BB_MIN                             // ASSOC GrpTRESMins exceeded(burst buffer)
	WAIT_ASSOC_GRP_BB_RUN_MIN                         // ASSOC GrpTRESRunMins exceeded(burst buffer)
	WAIT_ASSOC_MAX_BB_PER_JOB                         // ASSOC MaxTRESPerJob exceeded(burst buffer)
	WAIT_ASSOC_MAX_BB_PER_NODE                        // ASSOC MaxTRESPerNode exceeded(burst buffer)
	WAIT_ASSOC_MAX_BB_MINS_PER_JOB                    // ASSOC MaxTRESMinsPerJob exceeded(burst buffer)
	WAIT_QOS_GRP_UNK                                  // QOS GrpTRES exceeded (Unknown)
	WAIT_QOS_GRP_UNK_MIN                              // QOS GrpTRESMins exceeded (Unknown)
	WAIT_QOS_GRP_UNK_RUN_MIN                          // QOS GrpTRESRunMins exceeded (Unknown)
	WAIT_QOS_MAX_UNK_PER_JOB                          // QOS MaxTRESPerJob exceeded (Unknown)
	WAIT_QOS_MAX_UNK_PER_NODE                         // QOS MaxTRESPerNode exceeded (Unknown)
	WAIT_QOS_MAX_UNK_PER_USER                         // QOS MaxTRESPerUser exceeded (Unknown)
	WAIT_QOS_MAX_UNK_MINS_PER_JOB                     // QOS MaxTRESMinsPerJobexceeded (Unknown)
	WAIT_QOS_MIN_UNK                                  // QOS MinTRESPerJob exceeded (Unknown)
	WAIT_QOS_MAX_CPU_PER_NODE                         // QOS MaxTRESPerNode exceeded (CPU)
	WAIT_QOS_GRP_MEM_MIN                              // QOS GrpTRESMins exceeded(Memory)
	WAIT_QOS_GRP_MEM_RUN_MIN                          // QOS GrpTRESRunMins exceeded(Memory)
	WAIT_QOS_MAX_MEM_MINS_PER_JOB                     // QOS MaxTRESMinsPerJobexceeded (Memory)
	WAIT_QOS_MAX_MEM_PER_JOB                          // QOS MaxTRESPerJob exceeded (CPU)
	WAIT_QOS_MAX_MEM_PER_NODE                         // QOS MaxTRESPerNode exceeded (MEM)
	WAIT_QOS_MAX_MEM_PER_USER                         // QOS MaxTRESPerUser exceeded (CPU)
	WAIT_QOS_MIN_MEM                                  // QOS MinTRESPerJob not reached (Memory)
	WAIT_QOS_GRP_ENERGY                               // QOS GrpTRES exceeded (Energy)
	WAIT_QOS_GRP_ENERGY_MIN                           // QOS GrpTRESMins exceeded (Energy)
	WAIT_QOS_GRP_ENERGY_RUN_MIN                       // QOS GrpTRESRunMins exceeded (Energy)
	WAIT_QOS_MAX_ENERGY_PER_JOB                       // QOS MaxTRESPerJob exceeded (Energy)
	WAIT_QOS_MAX_ENERGY_PER_NODE                      // QOS MaxTRESPerNode exceeded (Energy)
	WAIT_QOS_MAX_ENERGY_PER_USER                      // QOS MaxTRESPerUser exceeded (Energy)
	WAIT_QOS_MAX_ENERGY_MINS_PER_JOB                  // QOS MaxTRESMinsPerJobexceeded (Energy)
	WAIT_QOS_MIN_ENERGY                               // QOS MinTRESPerJob not reached (Energy)
	WAIT_QOS_GRP_NODE_MIN                             // QOS GrpTRESMins exceeded (Node)
	WAIT_QOS_GRP_NODE_RUN_MIN                         // QOS GrpTRESRunMins exceeded (Node)
	WAIT_QOS_MAX_NODE_MINS_PER_JOB                    // QOS MaxTRESMinsPerJobexceeded (Node)
	WAIT_QOS_MIN_NODE                                 // QOS MinTRESPerJob not reached (Node)
	WAIT_QOS_GRP_GRES                                 // QOS GrpTRES exceeded (GRES)
	WAIT_QOS_GRP_GRES_MIN                             // QOS GrpTRESMins exceeded (GRES)
	WAIT_QOS_GRP_GRES_RUN_MIN                         // QOS GrpTRESRunMins exceeded (GRES)
	WAIT_QOS_MAX_GRES_PER_JOB                         // QOS MaxTRESPerJob exceeded (GRES)
	WAIT_QOS_MAX_GRES_PER_NODE                        // QOS MaxTRESPerNode exceeded (GRES)
	WAIT_QOS_MAX_GRES_PER_USER                        // QOS MaxTRESPerUser exceeded(GRES)
	WAIT_QOS_MAX_GRES_MINS_PER_JOB                    // QOS MaxTRESMinsPerJobexceeded (GRES)
	WAIT_QOS_MIN_GRES                                 // QOS MinTRESPerJob not reached (CPU)
	WAIT_QOS_GRP_LIC                                  // QOS GrpTRES exceeded (license)
	WAIT_QOS_GRP_LIC_MIN                              // QOS GrpTRESMins exceeded (license)
	WAIT_QOS_GRP_LIC_RUN_MIN                          // QOS GrpTRESRunMins exceeded (license)
	WAIT_QOS_MAX_LIC_PER_JOB                          // QOS MaxTRESPerJob exceeded (license)
	WAIT_QOS_MAX_LIC_PER_USER                         // QOS MaxTRESPerUser exceeded (license)
	WAIT_QOS_MAX_LIC_MINS_PER_JOB                     // QOS MaxTRESMinsPerJob exceeded(license)
	WAIT_QOS_MIN_LIC                                  // QOS MinTRESPerJob not reached(license)
	WAIT_QOS_GRP_BB                                   // QOS GrpTRES exceeded(burst buffer)
	WAIT_QOS_GRP_BB_MIN                               // QOS GrpTRESMins exceeded(burst buffer)
	WAIT_QOS_GRP_BB_RUN_MIN                           // QOS GrpTRESRunMins exceeded(burst buffer)
	WAIT_QOS_MAX_BB_PER_JOB                           // QOS MaxTRESPerJob exceeded(burst buffer)
	WAIT_QOS_MAX_BB_PER_NODE                          // QOS MaxTRESPerNode exceeded(burst buffer)
	WAIT_QOS_MAX_BB_PER_USER                          // QOS MaxTRESPerUser exceeded(burst buffer)
	WAIT_QOS_MAX_BB_MINS_PER_JOB                      // QOS MaxTRESMinsPerJob exceeded(burst buffer)
	WAIT_QOS_MIN_BB                                   // QOS MinTRESPerJob not reached(burst buffer)
	FAIL_DEADLINE                                     // reached deadline
	WAIT_QOS_MAX_BB_PER_ACCT                          // exceeded burst buffer
	WAIT_QOS_MAX_CPU_PER_ACCT                         // exceeded CPUs
	WAIT_QOS_MAX_ENERGY_PER_ACCT                      // exceeded Energy
	WAIT_QOS_MAX_GRES_PER_ACCT                        // exceeded GRES
	WAIT_QOS_MAX_NODE_PER_ACCT                        // exceeded Nodes
	WAIT_QOS_MAX_LIC_PER_ACCT                         // exceeded Licenses
	WAIT_QOS_MAX_MEM_PER_ACCT                         // exceeded Memory
	WAIT_QOS_MAX_UNK_PER_ACCT                         // exceeded Unknown
	WAIT_QOS_MAX_JOB_PER_ACCT                         // QOS MaxJobPerAccount exceeded
	WAIT_QOS_MAX_SUB_JOB_PER_ACCT                     // QOS MaxJobSubmitSPerAccount exceeded
	WAIT_PART_CONFIG                                  // Generic partition configuration reason
	WAIT_ACCOUNT_POLICY                               // Generic accounting policy reason
	WAIT_FED_JOB_LOCK                                 // Can't get fed job lock
	FAIL_OOM                                          // Exhausted memory
	WAIT_PN_MEM_LIMIT                                 // MaxMemPer[CPU|Node] exceeded
	WAIT_ASSOC_GRP_BILLING                            // GrpTRES
	WAIT_ASSOC_GRP_BILLING_MIN                        // GrpTRESMins
	WAIT_ASSOC_GRP_BILLING_RUN_MIN                    // GrpTRESRunMins
	WAIT_ASSOC_MAX_BILLING_PER_JOB                    // MaxTRESPerJob
	WAIT_ASSOC_MAX_BILLING_PER_NODE                   // MaxTRESPerNode
	WAIT_ASSOC_MAX_BILLING_MINS_PER_JOB               // MaxTRESMinsPerJob
	WAIT_QOS_GRP_BILLING                              // GrpTRES
	WAIT_QOS_GRP_BILLING_MIN                          // GrpTRESMins
	WAIT_QOS_GRP_BILLING_RUN_MIN                      // GrpTRESRunMins
	WAIT_QOS_MAX_BILLING_PER_JOB                      // MaxTRESPerJob
	WAIT_QOS_MAX_BILLING_PER_NODE                     // MaxTRESPerNode
	WAIT_QOS_MAX_BILLING_PER_USER                     // MaxTRESPerUser
	WAIT_QOS_MAX_BILLING_MINS_PER_JOB                 // MaxTRESMinsPerJob
	WAIT_QOS_MAX_BILLING_PER_ACCT                     // MaxTRESPerAcct
	WAIT_QOS_MIN_BILLING                              // MinTRESPerJob
	WAIT_RESV_DELETED                                 // Reservation was deleted
	WAIT_RESV_INVALID
	FAIL_CONSTRAINTS                       // Constraints cannot currently be satisfied
	WAIT_QOS_MAX_BB_RUN_MINS_PER_ACCT      // QOS MaxTRESRunMinsPerAccountexceeded (burst buffer)
	WAIT_QOS_MAX_BILLING_RUN_MINS_PER_ACCT // QOS MaxTRESRunMinsPerAccountexceeded (billing)
	WAIT_QOS_MAX_CPU_RUN_MINS_PER_ACCT     // QOS MaxTRESRunMinsPerAccountexceeded (CPU)
	WAIT_QOS_MAX_ENERGY_RUN_MINS_PER_ACCT  // QOS MaxTRESRunMinsPerAccountexceeded (Energy)
	WAIT_QOS_MAX_GRES_RUN_MINS_PER_ACCT    // QOS MaxTRESRunMinsPerAccountexceeded (GRES)
	WAIT_QOS_MAX_NODE_RUN_MINS_PER_ACCT    // QOS MaxTRESRunMinsPerAccountexceeded (Node)
	WAIT_QOS_MAX_LIC_RUN_MINS_PER_ACCT     // QOS MaxTRESRunMinsPerAccountexceeded (license)
	WAIT_QOS_MAX_MEM_RUN_MINS_PER_ACCT     // QOS MaxTRESRunMinsPerAccountexceeded (Memory)
	WAIT_QOS_MAX_UNK_RUN_MINS_PER_ACCT     // QOS MaxTRESRunMinsPerAccountexceeded (Unknown)
	WAIT_QOS_MAX_BB_RUN_MINS_PER_USER      // QOS MaxTRESRunMinsPerUser exceeded(burst buffer)
	WAIT_QOS_MAX_BILLING_RUN_MINS_PER_USER // QOS MaxTRESRunMinsPerUserexceeded (billing)
	WAIT_QOS_MAX_CPU_RUN_MINS_PER_USER     // QOS MaxTRESRunMinsPerUser exceeded(CPU)
	WAIT_QOS_MAX_ENERGY_RUN_MINS_PER_USER  // QOS MaxTRESRunMinsPerUserexceeded (Energy)
	WAIT_QOS_MAX_GRES_RUN_MINS_PER_USER    // QOS MaxTRESRunMinsPerUserexceeded (GRES)
	WAIT_QOS_MAX_NODE_RUN_MINS_PER_USER    // QOS MaxTRESRunMinsPerUserexceeded (Node)
	WAIT_QOS_MAX_LIC_RUN_MINS_PER_USER     // QOS MaxTRESRunMinsPerUser exceeded(license)
	WAIT_QOS_MAX_MEM_RUN_MINS_PER_USER     // QOS MaxTRESRunMinsPerUser exceeded(Memory)
	WAIT_QOS_MAX_UNK_RUN_MINS_PER_USER     // QOS MaxTRESRunMinsPerUser exceeded (Unknown)
	WAIT_MAX_POWERED_NODES                 //max_powered_nodes reached
	WAIT_MPI_PORTS_BUSY                    // MPI resv_ports busy
	REASON_END                             // end of table
)

var jsra = map[uint32]string{
	WAIT_NO_REASON:                         "None",
	WAIT_PROLOG:                            "Prolog",
	WAIT_PRIORITY:                          "Priority",
	WAIT_DEPENDENCY:                        "Dependency",
	WAIT_RESOURCES:                         "Resources",
	WAIT_PART_NODE_LIMIT:                   "PartitionNodeLimit",
	WAIT_PART_TIME_LIMIT:                   "PartitionTimeLimit",
	WAIT_PART_DOWN:                         "PartitionDown",
	WAIT_PART_INACTIVE:                     "PartitionInactive",
	WAIT_HELD:                              "JobHeldAdmin",
	WAIT_HELD_USER:                         "JobHeldUser",
	WAIT_TIME:                              "BeginTime",
	WAIT_LICENSES:                          "Licenses",
	WAIT_ASSOC_JOB_LIMIT:                   "AssociationJobLimit",
	WAIT_ASSOC_RESOURCE_LIMIT:              "AssociationResourceLimit",
	WAIT_ASSOC_TIME_LIMIT:                  "AssociationTimeLimit",
	WAIT_RESERVATION:                       "Reservation",
	WAIT_NODE_NOT_AVAIL:                    "ReqNodeNotAvail",
	FAIL_DEFER:                             "SchedDefer",
	FAIL_DOWN_PARTITION:                    "PartitionDown",
	FAIL_DOWN_NODE:                         "NodeDown",
	FAIL_BAD_CONSTRAINTS:                   "BadConstraints",
	FAIL_SYSTEM:                            "SystemFailure",
	FAIL_LAUNCH:                            "JobLaunchFailure",
	FAIL_EXIT_CODE:                         "NonZeroExitCode",
	FAIL_SIGNAL:                            "RaisedSignal",
	FAIL_TIMEOUT:                           "TimeLimit",
	FAIL_INACTIVE_LIMIT:                    "InactiveLimit",
	FAIL_ACCOUNT:                           "InvalidAccount",
	FAIL_QOS:                               "InvalidQOS",
	WAIT_QOS_THRES:                         "QOSUsageThreshold",
	WAIT_QOS_JOB_LIMIT:                     "QOSJobLimit",
	WAIT_QOS_RESOURCE_LIMIT:                "QOSResourceLimit",
	WAIT_QOS_TIME_LIMIT:                    "QOSTimeLimit",
	WAIT_CLEANING:                          "Cleaning",
	WAIT_QOS:                               "QOSNotAllowed",
	WAIT_ACCOUNT:                           "AccountNotAllowed",
	WAIT_DEP_INVALID:                       "DependencyNeverSatisfied",
	WAIT_QOS_GRP_CPU:                       "QOSGrpCpuLimit",
	WAIT_QOS_GRP_CPU_MIN:                   "QOSGrpCPUMinutesLimit",
	WAIT_QOS_GRP_CPU_RUN_MIN:               "QOSGrpCPURunMinutesLimit",
	WAIT_QOS_GRP_JOB:                       "QOSGrpJobsLimit",
	WAIT_QOS_GRP_MEM:                       "QOSGrpMemLimit",
	WAIT_QOS_GRP_NODE:                      "QOSGrpNodeLimit",
	WAIT_QOS_GRP_SUB_JOB:                   "QOSGrpSubmitJobsLimit",
	WAIT_QOS_GRP_WALL:                      "QOSGrpWallLimit",
	WAIT_QOS_MAX_CPU_PER_JOB:               "QOSMaxCpuPerJobLimit",
	WAIT_QOS_MAX_CPU_MINS_PER_JOB:          "QOSMaxCpuMinutesPerJobLimit",
	WAIT_QOS_MAX_NODE_PER_JOB:              "QOSMaxNodePerJobLimit",
	WAIT_QOS_MAX_WALL_PER_JOB:              "QOSMaxWallDurationPerJobLimit",
	WAIT_QOS_MAX_CPU_PER_USER:              "QOSMaxCpuPerUserLimit",
	WAIT_QOS_MAX_JOB_PER_USER:              "QOSMaxJobsPerUserLimit",
	WAIT_QOS_MAX_NODE_PER_USER:             "QOSMaxNodePerUserLimit",
	WAIT_QOS_MAX_SUB_JOB:                   "QOSMaxSubmitJobPerUserLimit",
	WAIT_QOS_MIN_CPU:                       "QOSMinCpuNotSatisfied",
	WAIT_ASSOC_GRP_CPU:                     "AssocGrpCpuLimit",
	WAIT_ASSOC_GRP_CPU_MIN:                 "AssocGrpCPUMinutesLimit",
	WAIT_ASSOC_GRP_CPU_RUN_MIN:             "AssocGrpCPURunMinutesLimit",
	WAIT_ASSOC_GRP_JOB:                     "AssocGrpJobsLimit",
	WAIT_ASSOC_GRP_MEM:                     "AssocGrpMemLimit",
	WAIT_ASSOC_GRP_NODE:                    "AssocGrpNodeLimit",
	WAIT_ASSOC_GRP_SUB_JOB:                 "AssocGrpSubmitJobsLimit",
	WAIT_ASSOC_GRP_WALL:                    "AssocGrpWallLimit",
	WAIT_ASSOC_MAX_JOBS:                    "AssocMaxJobsLimit",
	WAIT_ASSOC_MAX_CPU_PER_JOB:             "AssocMaxCpuPerJobLimit",
	WAIT_ASSOC_MAX_CPU_MINS_PER_JOB:        "AssocMaxCpuMinutesPerJobLimit",
	WAIT_ASSOC_MAX_NODE_PER_JOB:            "AssocMaxNodePerJobLimit",
	WAIT_ASSOC_MAX_WALL_PER_JOB:            "AssocMaxWallDurationPerJobLimit",
	WAIT_ASSOC_MAX_SUB_JOB:                 "AssocMaxSubmitJobLimit",
	WAIT_MAX_REQUEUE:                       "JobHoldMaxRequeue",
	WAIT_ARRAY_TASK_LIMIT:                  "JobArrayTaskLimit",
	WAIT_BURST_BUFFER_RESOURCE:             "BurstBufferResources",
	WAIT_BURST_BUFFER_STAGING:              "BurstBufferStageIn",
	FAIL_BURST_BUFFER_OP:                   "BurstBufferOperation",
	WAIT_ASSOC_GRP_UNK:                     "AssocGrpUnknown",
	WAIT_ASSOC_GRP_UNK_MIN:                 "AssocGrpUnknownMinutes",
	WAIT_ASSOC_GRP_UNK_RUN_MIN:             "AssocGrpUnknownRunMinutes",
	WAIT_ASSOC_MAX_UNK_PER_JOB:             "AssocMaxUnknownPerJob",
	WAIT_ASSOC_MAX_UNK_PER_NODE:            "AssocMaxUnknownPerNode",
	WAIT_ASSOC_MAX_UNK_MINS_PER_JOB:        "AssocMaxUnknownMinutesPerJob",
	WAIT_ASSOC_MAX_CPU_PER_NODE:            "AssocMaxCpuPerNode",
	WAIT_ASSOC_GRP_MEM_MIN:                 "AssocGrpMemMinutes",
	WAIT_ASSOC_GRP_MEM_RUN_MIN:             "AssocGrpMemRunMinutes",
	WAIT_ASSOC_MAX_MEM_PER_JOB:             "AssocMaxMemPerJob",
	WAIT_ASSOC_MAX_MEM_PER_NODE:            "AssocMaxMemPerNode",
	WAIT_ASSOC_MAX_MEM_MINS_PER_JOB:        "AssocMaxMemMinutesPerJob",
	WAIT_ASSOC_GRP_NODE_MIN:                "AssocGrpNodeMinutes",
	WAIT_ASSOC_GRP_NODE_RUN_MIN:            "AssocGrpNodeRunMinutes",
	WAIT_ASSOC_MAX_NODE_MINS_PER_JOB:       "AssocMaxNodeMinutesPerJob",
	WAIT_ASSOC_GRP_ENERGY:                  "AssocGrpEnergy",
	WAIT_ASSOC_GRP_ENERGY_MIN:              "AssocGrpEnergyMinutes",
	WAIT_ASSOC_GRP_ENERGY_RUN_MIN:          "AssocGrpEnergyRunMinutes",
	WAIT_ASSOC_MAX_ENERGY_PER_JOB:          "AssocMaxEnergyPerJob",
	WAIT_ASSOC_MAX_ENERGY_PER_NODE:         "AssocMaxEnergyPerNode",
	WAIT_ASSOC_MAX_ENERGY_MINS_PER_JOB:     "AssocMaxEnergyMinutesPerJob",
	WAIT_ASSOC_GRP_GRES:                    "AssocGrpGRES",
	WAIT_ASSOC_GRP_GRES_MIN:                "AssocGrpGRESMinutes",
	WAIT_ASSOC_GRP_GRES_RUN_MIN:            "AssocGrpGRESRunMinutes",
	WAIT_ASSOC_MAX_GRES_PER_JOB:            "AssocMaxGRESPerJob",
	WAIT_ASSOC_MAX_GRES_PER_NODE:           "AssocMaxGRESPerNode",
	WAIT_ASSOC_MAX_GRES_MINS_PER_JOB:       "AssocMaxGRESMinutesPerJob",
	WAIT_ASSOC_GRP_LIC:                     "AssocGrpLicense",
	WAIT_ASSOC_GRP_LIC_MIN:                 "AssocGrpLicenseMinutes",
	WAIT_ASSOC_GRP_LIC_RUN_MIN:             "AssocGrpLicenseRunMinutes",
	WAIT_ASSOC_MAX_LIC_PER_JOB:             "AssocMaxLicensePerJob",
	WAIT_ASSOC_MAX_LIC_MINS_PER_JOB:        "AssocMaxLicenseMinutesPerJob",
	WAIT_ASSOC_GRP_BB:                      "AssocGrpBB",
	WAIT_ASSOC_GRP_BB_MIN:                  "AssocGrpBBMinutes",
	WAIT_ASSOC_GRP_BB_RUN_MIN:              "AssocGrpBBRunMinutes",
	WAIT_ASSOC_MAX_BB_PER_JOB:              "AssocMaxBBPerJob",
	WAIT_ASSOC_MAX_BB_PER_NODE:             "AssocMaxBBPerNode",
	WAIT_ASSOC_MAX_BB_MINS_PER_JOB:         "AssocMaxBBMinutesPerJob",
	WAIT_QOS_GRP_UNK:                       "QOSGrpUnknown",
	WAIT_QOS_GRP_UNK_MIN:                   "QOSGrpUnknownMinutes",
	WAIT_QOS_GRP_UNK_RUN_MIN:               "QOSGrpUnknownRunMinutes",
	WAIT_QOS_MAX_UNK_PER_JOB:               "QOSMaxUnknownPerJob",
	WAIT_QOS_MAX_UNK_PER_NODE:              "QOSMaxUnknownPerNode",
	WAIT_QOS_MAX_UNK_PER_USER:              "QOSMaxUnknownPerUser",
	WAIT_QOS_MAX_UNK_MINS_PER_JOB:          "QOSMaxUnknownMinutesPerJob",
	WAIT_QOS_MIN_UNK:                       "QOSMinUnknown",
	WAIT_QOS_MAX_CPU_PER_NODE:              "QOSMaxCpuPerNode",
	WAIT_QOS_GRP_MEM_MIN:                   "QOSGrpMemoryMinutes",
	WAIT_QOS_GRP_MEM_RUN_MIN:               "QOSGrpMemoryRunMinutes",
	WAIT_QOS_MAX_MEM_PER_JOB:               "QOSMaxMemoryPerJob",
	WAIT_QOS_MAX_MEM_PER_NODE:              "QOSMaxMemoryPerNode",
	WAIT_QOS_MAX_MEM_PER_USER:              "QOSMaxMemoryPerUser",
	WAIT_QOS_MAX_MEM_MINS_PER_JOB:          "QOSMaxMemoryMinutesPerJob",
	WAIT_QOS_MIN_MEM:                       "QOSMinMemory",
	WAIT_QOS_GRP_NODE_MIN:                  "QOSGrpNodeMinutes",
	WAIT_QOS_GRP_NODE_RUN_MIN:              "QOSGrpNodeRunMinutes",
	WAIT_QOS_MAX_NODE_MINS_PER_JOB:         "QOSMaxNodeMinutesPerJob",
	WAIT_QOS_MIN_NODE:                      "QOSMinNode",
	WAIT_QOS_GRP_ENERGY:                    "QOSGrpEnergy",
	WAIT_QOS_GRP_ENERGY_MIN:                "QOSGrpEnergyMinutes",
	WAIT_QOS_GRP_ENERGY_RUN_MIN:            "QOSGrpEnergyRunMinutes",
	WAIT_QOS_MAX_ENERGY_PER_JOB:            "QOSMaxEnergyPerJob",
	WAIT_QOS_MAX_ENERGY_PER_NODE:           "QOSMaxEnergyPerNode",
	WAIT_QOS_MAX_ENERGY_PER_USER:           "QOSMaxEnergyPerUser",
	WAIT_QOS_MAX_ENERGY_MINS_PER_JOB:       "QOSMaxEnergyMinutesPerJob",
	WAIT_QOS_MIN_ENERGY:                    "QOSMinEnergy",
	WAIT_QOS_GRP_GRES:                      "QOSGrpGRES",
	WAIT_QOS_GRP_GRES_MIN:                  "QOSGrpGRESMinutes",
	WAIT_QOS_GRP_GRES_RUN_MIN:              "QOSGrpGRESRunMinutes",
	WAIT_QOS_MAX_GRES_PER_JOB:              "QOSMaxGRESPerJob",
	WAIT_QOS_MAX_GRES_PER_NODE:             "QOSMaxGRESPerNode",
	WAIT_QOS_MAX_GRES_PER_USER:             "QOSMaxGRESPerUser",
	WAIT_QOS_MAX_GRES_MINS_PER_JOB:         "QOSMaxGRESMinutesPerJob",
	WAIT_QOS_MIN_GRES:                      "QOSMinGRES",
	WAIT_QOS_GRP_LIC:                       "QOSGrpLicense",
	WAIT_QOS_GRP_LIC_MIN:                   "QOSGrpLicenseMinutes",
	WAIT_QOS_GRP_LIC_RUN_MIN:               "QOSGrpLicenseRunMinutes",
	WAIT_QOS_MAX_LIC_PER_JOB:               "QOSMaxLicensePerJob",
	WAIT_QOS_MAX_LIC_PER_USER:              "QOSMaxLicensePerUser",
	WAIT_QOS_MAX_LIC_MINS_PER_JOB:          "QOSMaxLicenseMinutesPerJob",
	WAIT_QOS_MIN_LIC:                       "QOSMinLicense",
	WAIT_QOS_GRP_BB:                        "QOSGrpBB",
	WAIT_QOS_GRP_BB_MIN:                    "QOSGrpBBMinutes",
	WAIT_QOS_GRP_BB_RUN_MIN:                "QOSGrpBBRunMinutes",
	WAIT_QOS_MAX_BB_PER_JOB:                "QOSMaxBBPerJob",
	WAIT_QOS_MAX_BB_PER_NODE:               "QOSMaxBBPerNode",
	WAIT_QOS_MAX_BB_PER_USER:               "QOSMaxBBPerUser",
	WAIT_QOS_MAX_BB_MINS_PER_JOB:           "AssocMaxBBMinutesPerJob",
	WAIT_QOS_MIN_BB:                        "QOSMinBB",
	FAIL_DEADLINE:                          "DeadLine",
	WAIT_QOS_MAX_BB_PER_ACCT:               "MaxBBPerAccount",
	WAIT_QOS_MAX_CPU_PER_ACCT:              "MaxCpuPerAccount",
	WAIT_QOS_MAX_ENERGY_PER_ACCT:           "MaxEnergyPerAccount",
	WAIT_QOS_MAX_GRES_PER_ACCT:             "MaxGRESPerAccount",
	WAIT_QOS_MAX_NODE_PER_ACCT:             "MaxNodePerAccount",
	WAIT_QOS_MAX_LIC_PER_ACCT:              "MaxLicensePerAccount",
	WAIT_QOS_MAX_MEM_PER_ACCT:              "MaxMemoryPerAccount",
	WAIT_QOS_MAX_UNK_PER_ACCT:              "MaxUnknownPerAccount",
	WAIT_QOS_MAX_JOB_PER_ACCT:              "MaxJobsPerAccount",
	WAIT_QOS_MAX_SUB_JOB_PER_ACCT:          "MaxSubmitJobsPerAccount",
	WAIT_PART_CONFIG:                       "PartitionConfig",
	WAIT_ACCOUNT_POLICY:                    "AccountingPolicy",
	WAIT_FED_JOB_LOCK:                      "FedJobLock",
	FAIL_OOM:                               "OutOfMemory",
	WAIT_PN_MEM_LIMIT:                      "MaxMemPerLimit",
	WAIT_ASSOC_GRP_BILLING:                 "AssocGrpBilling",
	WAIT_ASSOC_GRP_BILLING_MIN:             "AssocGrpBillingMinutes",
	WAIT_ASSOC_GRP_BILLING_RUN_MIN:         "AssocGrpBillingRunMinutes",
	WAIT_ASSOC_MAX_BILLING_PER_JOB:         "AssocMaxBillingPerJob",
	WAIT_ASSOC_MAX_BILLING_PER_NODE:        "AssocMaxBillingPerNode",
	WAIT_ASSOC_MAX_BILLING_MINS_PER_JOB:    "AssocMaxBillingMinutesPerJob",
	WAIT_QOS_GRP_BILLING:                   "QOSGrpBilling",
	WAIT_QOS_GRP_BILLING_MIN:               "QOSGrpBillingMinutes",
	WAIT_QOS_GRP_BILLING_RUN_MIN:           "QOSGrpBillingRunMinutes",
	WAIT_QOS_MAX_BILLING_PER_JOB:           "QOSMaxBillingPerJob",
	WAIT_QOS_MAX_BILLING_PER_NODE:          "QOSMaxBillingPerNode",
	WAIT_QOS_MAX_BILLING_PER_USER:          "QOSMaxBillingPerUser",
	WAIT_QOS_MAX_BILLING_MINS_PER_JOB:      "QOSMaxBillingMinutesPerJob",
	WAIT_QOS_MAX_BILLING_PER_ACCT:          "MaxBillingPerAccount",
	WAIT_QOS_MIN_BILLING:                   "QOSMinBilling",
	WAIT_RESV_DELETED:                      "ReservationDeleted",
	WAIT_RESV_INVALID:                      "ReservationInvalid",
	FAIL_CONSTRAINTS:                       "Constraints",
	WAIT_QOS_MAX_BB_RUN_MINS_PER_ACCT:      "MaxBBRunMinsPerAccount",
	WAIT_QOS_MAX_BILLING_RUN_MINS_PER_ACCT: "MaxBillingRunMinsPerAccount",
	WAIT_QOS_MAX_CPU_RUN_MINS_PER_ACCT:     "MaxCpuRunMinsPerAccount",
	WAIT_QOS_MAX_ENERGY_RUN_MINS_PER_ACCT:  "MaxEnergyRunMinsPerAccount",
	WAIT_QOS_MAX_GRES_RUN_MINS_PER_ACCT:    "MaxGRESRunMinsPerAccount",
	WAIT_QOS_MAX_NODE_RUN_MINS_PER_ACCT:    "MaxNodeRunMinsPerAccount",
	WAIT_QOS_MAX_LIC_RUN_MINS_PER_ACCT:     "MaxLicenseRunMinsPerAccount",
	WAIT_QOS_MAX_MEM_RUN_MINS_PER_ACCT:     "MaxMemoryRunMinsPerAccount",
	WAIT_QOS_MAX_UNK_RUN_MINS_PER_ACCT:     "MaxUnknownRunMinsPerAccount",
	WAIT_QOS_MAX_BB_RUN_MINS_PER_USER:      "MaxBBRunMinsPerUser",
	WAIT_QOS_MAX_BILLING_RUN_MINS_PER_USER: "MaxBillingRunMinsPerUser",
	WAIT_QOS_MAX_CPU_RUN_MINS_PER_USER:     "MaxCpuRunMinsPerUser",
	WAIT_QOS_MAX_ENERGY_RUN_MINS_PER_USER:  "MaxEnergyRunMinsPerUser",
	WAIT_QOS_MAX_GRES_RUN_MINS_PER_USER:    "MaxGRESRunMinsPerUser",
	WAIT_QOS_MAX_NODE_RUN_MINS_PER_USER:    "MaxNodeRunMinsPerUser",
	WAIT_QOS_MAX_LIC_RUN_MINS_PER_USER:     "MaxLicenseRunMinsPerUser",
	WAIT_QOS_MAX_MEM_RUN_MINS_PER_USER:     "MaxMemoryRunMinsPerUser",
	WAIT_QOS_MAX_UNK_RUN_MINS_PER_USER:     "MaxUnknownRunMinsPerUser",
	WAIT_MAX_POWERED_NODES:                 "MaxPoweredUpNodes",
	WAIT_MPI_PORTS_BUSY:                    "MpiPortsBusy",
}

// 将 state_reason_prev 输出为 string.
func PrintJobStateReasonStr(ireason uint32) string {
	reason := "InvalidReason"
	if str, ok := jsra[ireason]; (ireason < REASON_END) && ok {
		return str
	}
	return reason
}
