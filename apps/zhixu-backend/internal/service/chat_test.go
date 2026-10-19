package service

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChatInputAndUserIsolation(t *testing.T) {
	s := newAIConfigTestService(t)
	view, err := s.SaveProvider(context.Background(), 7, 0, testProviderInput())
	require.NoError(t, err)
	chat := &ChatService{Store: s.Store, Credentials: s.Credentials}
	input := ChatInput{ModelID: view.Providers[0].Models[0].ID, Messages: []ChatMessage{{Role: "user", Content: "你好"}}}
	cap, err := chat.Resolve(context.Background(), 7, input)
	require.NoError(t, err)
	require.Equal(t, "test-secret-not-for-production", cap.Key)
	_, err = chat.Resolve(context.Background(), 8, input)
	require.ErrorIs(t, err, ErrAIConfigInvalid)
	input.ModelID = view.Providers[0].Models[1].ID
	_, err = chat.Resolve(context.Background(), 7, input)
	require.ErrorIs(t, err, ErrAIConfigInvalid)
	for _, msgs := range [][]ChatMessage{nil, {{Role: "system", Content: "x"}}, {{Role: "user", Content: " "}}, {{Role: "assistant", Content: "x"}}, {{Role: "user", Content: "x"}, {Role: "assistant", Content: "y"}}} {
		require.Error(t, ValidateChat(ChatInput{Messages: msgs}))
	}
}
func TestChatCapabilityLifetime(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	s := &ChatService{}
	cap := &ChatCapability{Context: ctx}
	token, revoke, err := s.Register(cap)
	require.NoError(t, err)
	require.Len(t, token, 64)
	require.Equal(t, cap, s.Capability(token))
	cancel()
	require.Nil(t, s.Capability(token))
	revoke()
	require.Nil(t, s.Capability(token))
}
func TestChatPrivateEndpointPolicy(t *testing.T) {
	for _, ip := range []string{"127.0.0.1", "::1", "10.0.0.1", "169.254.169.254", "100.100.100.200", "100.64.0.1"} {
		require.True(t, unsafeModelIP(net.ParseIP(ip)))
	}
	require.False(t, unsafeModelIP(net.ParseIP("8.8.8.8")))
	s := &ChatService{}
	_, _, err := s.ModelClient(context.Background(), "http://127.0.0.1:11434/v1")
	require.Error(t, err)
	s.AllowedPrivateEndpoints = "http://127.0.0.1:11434/v1"
	_, cleanup, err := s.ModelClient(context.Background(), "http://127.0.0.1:11434/v1")
	require.NoError(t, err)
	cleanup()
	_, _, err = s.ModelClient(context.Background(), "http://127.0.0.1:11435/v1")
	require.Error(t, err)
}
