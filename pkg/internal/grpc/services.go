package grpc

import (
	"context"
	"git.solsynth.dev/hypernet/nexus/pkg/nex"
	"strconv"

	"git.solsynth.dev/hypernet/nexus/pkg/proto"
	"git.solsynth.dev/hypernet/wallet/pkg/internal/database"
)

func (v *Server) BroadcastEvent(ctx context.Context, in *proto.EventInfo) (*proto.EventResponse, error) {
	switch in.GetEvent() {
	case "deletion":
		data := nex.DecodeMap(in.GetData())
		resType, ok := data["type"].(string)
		if !ok {
			break
		}
		switch resType {
		case "account":
			id, ok := data["id"].(string)
			if !ok {
				break
			}
			numericId, err := strconv.Atoi(id)
			if err != nil {
				break
			}
			tx := database.C.Begin()
			for _, model := range database.AutoMaintainRange {
				switch model.(type) {
				default:
					tx.Delete(model, "account_id = ?", numericId)
				}
			}
			tx.Commit()
		}
	}

	return &proto.EventResponse{}, nil
}
