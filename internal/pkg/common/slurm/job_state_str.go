package slurm

const (
	JOB_PENDING   uint64 = iota // queued waiting for initiation
	JOB_RUNNING                 // allocated resources and executing
	JOB_SUSPENDED               // allocated resources, execution suspended
	JOB_COMPLETE                //completed execution successfully
	JOB_CANCELLED               // cancelled by user
	JOB_FAILED                  // completed execution unsuccessfully
	JOB_TIMEOUT                 //terminated on reaching time limit
	JOB_NODE_FAIL               // terminated on node failure
	JOB_PREEMPTED               // terminated due to preemption
	JOB_BOOT_FAIL               // terminated due to node boot failure
	JOB_DEADLINE                // terminated on deadline
	JOB_OOM                     // experienced out of memory error
	JOB_END                     // not a real state, last entry in table
)

var (
	JOB_STATE_BASE    uint64 = 0x000000ff
	JOB_REQUEUE              = slurmBit(10) // Requeue job in completing state
	JOB_REQUEUE_HOLD         = slurmBit(11) // Requeue any job in hold
	JOB_SPECIAL_EXIT         = slurmBit(12) // Requeue an exit job in hold
	JOB_RESIZING             = slurmBit(13) // Size of job about to change, flag setbefore calling accounting functionsimmediately before job changes size
	JOB_CONFIGURING          = slurmBit(14) // Allocated nodes booting
	JOB_COMPLETING           = slurmBit(15) // Waiting for epilog completion
	JOB_STOPPED              = slurmBit(16) // Job is stopped state (holding resources, but sent SIGSTOP
	JOB_REVOKED              = slurmBit(19) // Sibling job revoked
	JOB_REQUEUE_FED          = slurmBit(20) // Job being requeued by federation
	JOB_RESV_DEL_HOLD        = slurmBit(21) // Job is hold
	JOB_SIGNALING            = slurmBit(22) // Outgoing signal is pending
	JOB_STAGE_OUT            = slurmBit(23) //Staging out data (burst buffer)
)

func slurmBit(offset uint) uint64 {
	return uint64(1) << offset
}

func PrintJobStateString(istate uint64) string {
	// Process JOB_STATE_FLAGS
	if (istate & JOB_COMPLETING) != 0 {
		return "COMPLETING"
	}

	if (istate & JOB_STAGE_OUT) != 0 {
		return "STAGE_OUT"
	}

	if (istate & JOB_CONFIGURING) != 0 {
		return "CONFIGURING"
	}

	if (istate & JOB_RESIZING) != 0 {
		return "RESIZING"
	}

	if (istate & JOB_REQUEUE) != 0 {
		return "REQUEUED"
	}

	if (istate & JOB_REQUEUE_FED) != 0 {
		return "REQUEUE_FED"
	}

	if (istate & JOB_REQUEUE_HOLD) != 0 {
		return "REQUEUE_HOLD"
	}

	if (istate & JOB_SPECIAL_EXIT) != 0 {
		return "SPECIAL_EXIT"
	}

	if (istate & JOB_STOPPED) != 0 {
		return "STOPPED"
	}

	if (istate & JOB_REVOKED) != 0 {
		return "REVOKED"
	}

	if (istate & JOB_RESV_DEL_HOLD) != 0 {
		return "RESV_DEL_HOLD"
	}

	if (istate & JOB_SIGNALING) != 0 {
		return "SIGNALING"
	}

	switch istate & JOB_STATE_BASE {
	case JOB_PENDING:
		return "PENDING"
	case JOB_RUNNING:
		return "RUNNING"
	case JOB_SUSPENDED:
		return "SUSPENDED"
	case JOB_COMPLETE:
		return "COMPLETED"
	case JOB_CANCELLED:
		return "CANCELLED"
	case JOB_FAILED:
		return "FAILED"
	case JOB_TIMEOUT:
		return "TIMEOUT"
	case JOB_NODE_FAIL:
		return "NODE_FAIL"
	case JOB_PREEMPTED:
		return "PREEMPTED"
	case JOB_BOOT_FAIL:
		return "BOOT_FAIL"
	case JOB_DEADLINE:
		return "DEADLINE"
	case JOB_OOM:
		return "OUT_OF_MEMORY"
	default:
		return "?"
	}
}
