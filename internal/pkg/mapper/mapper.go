// Package mapper содержит функции преобразования доменных моделей
// (банковские карты, учётные данные, текстовые и бинарные данные)
// в protobuf-структуры и обратно.
package mapper

import (
	"github.com/ryabkov82/gophkeeper/internal/domain/model"
	pb "github.com/ryabkov82/gophkeeper/internal/pkg/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// BankCardToPB converts model.BankCard to pb.BankCard.
func BankCardToPB(c *model.BankCard) *pb.BankCard {
	card := &pb.BankCard{}
	card.SetId(c.ID)
	card.SetUserId(c.UserID)
	card.SetTitle(c.Title)
	card.SetCardholderName(c.CardholderName)
	card.SetCardNumber(c.CardNumber)
	card.SetExpiryDate(c.ExpiryDate)
	card.SetCvv(c.CVV)
	card.SetMetadata(c.Metadata)
	card.SetCreatedAt(timestamppb.New(c.CreatedAt))
	card.SetUpdatedAt(timestamppb.New(c.UpdatedAt))
	return card
}

// BankCardFromPB converts pb.BankCard to model.BankCard.
func BankCardFromPB(pbCard *pb.BankCard) *model.BankCard {
	if pbCard == nil {
		return nil
	}
	return &model.BankCard{
		ID:             pbCard.GetId(),
		UserID:         pbCard.GetUserId(),
		Title:          pbCard.GetTitle(),
		CardholderName: pbCard.GetCardholderName(),
		CardNumber:     pbCard.GetCardNumber(),
		ExpiryDate:     pbCard.GetExpiryDate(),
		CVV:            pbCard.GetCvv(),
		Metadata:       pbCard.GetMetadata(),
		CreatedAt:      pbCard.GetCreatedAt().AsTime(),
		UpdatedAt:      pbCard.GetUpdatedAt().AsTime(),
	}
}

// CredentialToPB converts model.Credential to pb.Credential.
func CredentialToPB(c *model.Credential) *pb.Credential {
	cred := &pb.Credential{}
	cred.SetId(c.ID)
	cred.SetUserId(c.UserID)
	cred.SetTitle(c.Title)
	cred.SetLogin(c.Login)
	cred.SetPassword(c.Password)
	cred.SetMetadata(c.Metadata)
	cred.SetCreatedAt(timestamppb.New(c.CreatedAt))
	cred.SetUpdatedAt(timestamppb.New(c.UpdatedAt))
	return cred
}

// CredentialFromPB converts pb.Credential to model.Credential.
func CredentialFromPB(pbCred *pb.Credential) *model.Credential {
	if pbCred == nil {
		return nil
	}
	return &model.Credential{
		ID:        pbCred.GetId(),
		UserID:    pbCred.GetUserId(),
		Title:     pbCred.GetTitle(),
		Login:     pbCred.GetLogin(),
		Password:  pbCred.GetPassword(),
		Metadata:  pbCred.GetMetadata(),
		CreatedAt: pbCred.GetCreatedAt().AsTime(),
		UpdatedAt: pbCred.GetUpdatedAt().AsTime(),
	}
}

// TextDataToPB converts model.TextData to pb.TextData.
func TextDataToPB(td *model.TextData) *pb.TextData {
	pbtd := &pb.TextData{}
	pbtd.SetId(td.ID)
	pbtd.SetUserId(td.UserID)
	pbtd.SetTitle(td.Title)
	pbtd.SetContent(td.Content)
	pbtd.SetMetadata(td.Metadata)
	pbtd.SetCreatedAt(timestamppb.New(td.CreatedAt))
	pbtd.SetUpdatedAt(timestamppb.New(td.UpdatedAt))
	return pbtd
}

// TextDataFromPB converts pb.TextData to model.TextData.
func TextDataFromPB(pbtd *pb.TextData) *model.TextData {
	if pbtd == nil {
		return nil
	}
	return &model.TextData{
		ID:        pbtd.GetId(),
		UserID:    pbtd.GetUserId(),
		Title:     pbtd.GetTitle(),
		Content:   pbtd.GetContent(),
		Metadata:  pbtd.GetMetadata(),
		CreatedAt: pbtd.GetCreatedAt().AsTime(),
		UpdatedAt: pbtd.GetUpdatedAt().AsTime(),
	}
}

// BinaryDataToPB converts model.BinaryData to pb.BinaryDataInfo.
func BinaryDataToPB(bd *model.BinaryData) *pb.BinaryDataInfo {
	if bd == nil {
		return nil
	}
	info := &pb.BinaryDataInfo{}
	info.SetId(bd.ID)
	info.SetTitle(bd.Title)
	info.SetMetadata(bd.Metadata)
	info.SetSize(bd.Size)
	info.SetClientPath(bd.ClientPath)
	info.SetCreatedAt(timestamppb.New(bd.CreatedAt))
	info.SetUpdatedAt(timestamppb.New(bd.UpdatedAt))
	return info
}

// BinaryDataFromPB converts pb.BinaryDataInfo to model.BinaryData.
func BinaryDataFromPB(info *pb.BinaryDataInfo) *model.BinaryData {
	if info == nil {
		return nil
	}
	return &model.BinaryData{
		ID:         info.GetId(),
		Title:      info.GetTitle(),
		Metadata:   info.GetMetadata(),
		Size:       info.GetSize(),
		ClientPath: info.GetClientPath(),
		CreatedAt:  info.GetCreatedAt().AsTime(),
		UpdatedAt:  info.GetUpdatedAt().AsTime(),
	}
}
