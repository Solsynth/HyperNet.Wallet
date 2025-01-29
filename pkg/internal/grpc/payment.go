package grpc

import (
	"context"

	"git.solsynth.dev/hypernet/wallet/pkg/internal/database"
	"git.solsynth.dev/hypernet/wallet/pkg/internal/models"
	"git.solsynth.dev/hypernet/wallet/pkg/internal/services"
	"git.solsynth.dev/hypernet/wallet/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (v *Server) GetWallet(ctx context.Context, request *proto.GetWalletRequest) (*proto.GetWalletResponse, error) {
	var wallet models.Wallet
	if err := database.C.Where("account_id = ?", request.AccountId).First(&wallet).Error; err != nil {
		return nil, status.Errorf(codes.NotFound, "wallet not found")
	}

	return &proto.GetWalletResponse{
		Wallet: wallet.ToWalletInfo(),
	}, nil
}

func (v *Server) GetTransaction(ctx context.Context, request *proto.GetTransactionRequest) (*proto.GetTransactionResponse, error) {
	var transaction models.Transaction
	if err := database.C.Where("id = ?", request.Id).First(&transaction).Error; err != nil {
		return nil, status.Errorf(codes.NotFound, "transaction not found")
	}

	return &proto.GetTransactionResponse{
		Transaction: transaction.ToTransactionInfo(),
	}, nil
}

func (v *Server) MakeTransaction(ctx context.Context, request *proto.MakeTransactionRequest) (*proto.TransactionInfo, error) {
	if request.PayerId == nil && request.PayeeId == nil {
		return nil, status.Errorf(codes.InvalidArgument, "payer and payee cannot be both nil")
	}

	var payerWallet, payeeWallet *models.Wallet
	if request.PayerId != nil {
		if err := database.C.Where("id = ?", request.PayerId).First(&payerWallet); err != nil {
			return nil, status.Errorf(codes.NotFound, "payer wallet not found: %v", err)
		}
	}
	if request.PayeeId != nil {
		if err := database.C.Where("id = ?", request.PayeeId).First(&payeeWallet); err != nil {
			return nil, status.Errorf(codes.NotFound, "payee wallet not found: %v", err)
		}
	}

	transaction, err := services.MakeTransaction(request.GetAmount(), request.GetRemark(), payerWallet, payeeWallet)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return transaction.ToTransactionInfo(), nil
}

func (v *Server) MakeTransactionWithAccount(ctx context.Context, request *proto.MakeTransactionWithAccountRequest) (*proto.TransactionInfo, error) {
	if request.PayerAccountId == nil && request.PayeeAccountId == nil {
		return nil, status.Errorf(codes.InvalidArgument, "payer and payee cannot be both nil")
	}

	var payerWallet, payeeWallet *models.Wallet
	if request.PayerAccountId != nil {
		val := uint(*request.PayerAccountId)
		if err := database.C.Where("account_id = ?", val).First(&payerWallet).Error; err != nil {
			return nil, status.Errorf(codes.NotFound, "payer wallet not found: %v", err)
		}
	}
	if request.PayeeAccountId != nil {
		val := uint(*request.PayeeAccountId)
		if err := database.C.Where("account_id = ?", val).First(&payeeWallet).Error; err != nil {
			return nil, status.Errorf(codes.NotFound, "payee wallet not found: %v", err)
		}
	}

	transaction, err := services.MakeTransaction(request.GetAmount(), request.GetRemark(), payerWallet, payeeWallet)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return transaction.ToTransactionInfo(), nil
}
