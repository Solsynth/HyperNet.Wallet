package grpc

import (
	"context"
	"git.solsynth.dev/hypernet/wallet/pkg/internal/database"
	"git.solsynth.dev/hypernet/wallet/pkg/internal/models"
	"git.solsynth.dev/hypernet/wallet/pkg/proto"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
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

	transaction := models.Transaction{
		Amount: decimal.NewFromFloat(request.Amount),
		Remark: request.Remark,
	}
	if request.PayerId != nil {
		transaction.PayerID = lo.ToPtr(uint(*request.PayerId))
	}
	if request.PayeeId != nil {
		transaction.PayeeID = lo.ToPtr(uint(*request.PayeeId))
	}

	if err := database.C.Create(&transaction).Error; err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return transaction.ToTransactionInfo(), nil
}

func (v *Server) MakeTransactionWithAccount(ctx context.Context, request *proto.MakeTransactionWithAccountRequest) (*proto.TransactionInfo, error) {
	if request.PayerAccountId == nil && request.PayeeAccountId == nil {
		return nil, status.Errorf(codes.InvalidArgument, "payer and payee cannot be both nil")
	}

	transaction := models.Transaction{
		Amount: decimal.NewFromFloat(request.Amount),
		Remark: request.Remark,
	}
	if request.PayerAccountId != nil {
		val := uint(*request.PayerAccountId)
		var wallet models.Wallet
		if err := database.C.Where("account_id = ?", val).First(&wallet).Error; err != nil {
			return nil, status.Errorf(codes.NotFound, "payer wallet not found")
		}
		transaction.Payer = &wallet
		transaction.PayerID = &wallet.ID
	}
	if request.PayeeAccountId != nil {
		val := uint(*request.PayeeAccountId)
		var wallet models.Wallet
		if err := database.C.Where("account_id = ?", val).First(&wallet).Error; err != nil {
			return nil, status.Errorf(codes.NotFound, "payee wallet not found")
		}
		transaction.Payee = &wallet
		transaction.PayeeID = &wallet.ID
	}

	if err := database.C.Create(&transaction).Error; err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return transaction.ToTransactionInfo(), nil
}
