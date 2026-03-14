// Package gcloudsecrets provides wrapper over [secretmanager.Client]
package gcloudsecrets

import (
	"context"
	"errors"
	"fmt"
	"hash/crc32"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

type Manager struct {
	client      *secretmanager.Client
	projectName string
}

func NewManager(ctx context.Context, projectName string) (*Manager, error) {
	cli, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &Manager{
		client:      cli,
		projectName: projectName,
	}, nil
}

func (m *Manager) Close() error {
	return m.client.Close()
}

func (m *Manager) GetSecret(ctx context.Context, name string) (string, error) {
	fullName := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", m.projectName, name)

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fullName,
	}

	res, err := m.client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret version: %w", err)
	}

	tab := crc32.MakeTable(crc32.Castagnoli)

	checksum := int64(crc32.Checksum(res.Payload.Data, tab))
	if checksum != res.Payload.GetDataCrc32C() {
		return "", errors.New("checksums do not match")
	}

	return string(res.Payload.GetData()), nil
}
