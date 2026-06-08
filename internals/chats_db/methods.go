package chats_db

import (
	"ai-service-go/internals/types"
	"context"
	"fmt"
)

func (db *ChatDB) AddChat(ctx context.Context, chat types.AnnaChatType) (string, error) {
	var id string
	err := db.Pool.QueryRow(
		ctx,
		`INSERT INTO anna_chats (user_name, query, response, feedback, input_tokens, output_token, model, created_at)
         VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
         RETURNING id`,
		chat.UserName, chat.Query, chat.Response, chat.Feedback, chat.InputTokens, chat.OutputTokens, chat.Model,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("error inserting chat: %v", err)
	}
	return id, nil
}

func (db *ChatDB) UpdateChatFeedback(ctx context.Context, chatID int, feedback int) error {
	_, err := db.Pool.Exec(ctx, `UPDATE anna_chats SET feedback = $1 WHERE id = $2`, feedback, chatID)
	if err != nil {
		return fmt.Errorf("error updating chat feedback: %v", err)
	}
	return nil
}

func (db *ChatDB) GetChatsForAUser(ctx context.Context, name string) (types.AnnaChatType, error) {
	var layerInfo types.AnnaChatType
	err := db.Pool.QueryRow(ctx, `SELECT user_name, query, response, feedback, input_tokens, output_token, model, created_at FROM anna_chats WHERE user_name = $1`, name).Scan(&layerInfo.UserName, &layerInfo.Query, &layerInfo.Response, &layerInfo.Feedback, &layerInfo.InputTokens, &layerInfo.OutputTokens, &layerInfo.Model, &layerInfo.CreatedAt)
	if err != nil {
		if err.Error() == "no rows in result set" {
			return types.AnnaChatType{}, fmt.Errorf("layer information not found for name: %s", name)
		}
		return types.AnnaChatType{}, fmt.Errorf("error retrieving layer information: %v", err)
	}
	return layerInfo, nil
}

func (db *ChatDB) GetAllChats(ctx context.Context) ([]types.AnnaChatType, error) {
	var allChats []types.AnnaChatType
	rows, err := db.Pool.Query(ctx, `SELECT user_name, query, response, feedback, input_tokens, output_token, model, created_at FROM anna_chats`)
	if err != nil {
		return nil, fmt.Errorf("error retrieving all layer information: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var chat types.AnnaChatType
		err := rows.Scan(&chat.UserName, &chat.Query, &chat.Response, &chat.Feedback, &chat.InputTokens, &chat.OutputTokens, &chat.Model, &chat.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning layer information: %v", err)
		}
		allChats = append(allChats, chat)
	}
	return allChats, nil
}
