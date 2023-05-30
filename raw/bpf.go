package raw

import (
	"github.com/kataras/golog"
	"golang.org/x/net/bpf"
)

type Filter []bpf.Instruction

// minPort <= port <= maxPort
func createFilter(minPort, maxPort int, tos int) []bpf.Instruction {
	golog.Infof("set filter for %d ~ %d, tos: %d", minPort, maxPort, tos)
	filter := Filter{
		bpf.LoadIndirect{Off: 9, Size: 1},
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: uint32(17), SkipFalse: 4}, //UDP proto = 17

		// bpf.LoadIndirect{Off: 1, Size: 1},
		// bpf.JumpIf{Cond: bpf.JumpEqual, Val: uint32(tos), SkipFalse: 4}, //tos

		bpf.LoadAbsolute{Off: 22, Size: 2}, // load the dest port
		bpf.JumpIf{Cond: bpf.JumpGreaterOrEqual, Val: uint32(minPort), SkipFalse: 2},
		bpf.JumpIf{Cond: bpf.JumpLessOrEqual, Val: uint32(maxPort), SkipFalse: 1},
		bpf.RetConstant{Val: 0xffff},
		bpf.RetConstant{Val: 0x0},
	}

	return filter
}

func createEmptyFilter() Filter {
	filter := Filter{
		bpf.RetConstant{Val: 0x0},
	}
	return filter
}
