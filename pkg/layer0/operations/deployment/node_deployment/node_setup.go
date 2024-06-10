package node_deployment

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os/exec"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"golang.org/x/crypto/argon2"
	"golang.org/x/net/context"
)

// NodeConfig holds the configuration for a blockchain node
type NodeConfig struct {
	NetworkID       string
	NodeID          string
	IP              string
	Port            int
	Consensus       string
	InitialNodes    []string
	ContainerConfig *ContainerConfig
}

// ContainerConfig holds the configuration for a containerized node
type ContainerConfig struct {
	Image       string
	Volumes     map[string]string
	Environment []string
}

// NodeManager handles the deployment, management, and monitoring of nodes
type NodeManager struct {
	client *client.Client
	ctx    context.Context
}

// NewNodeManager initializes a new NodeManager
func NewNodeManager() (*NodeManager, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %v", err)
	}
	cli.NegotiateAPIVersion(context.Background())

	return &NodeManager{
		client: cli,
		ctx:    context.Background(),
	}, nil
}

// DeployNode deploys a new blockchain node
func (nm *NodeManager) DeployNode(config NodeConfig) (string, error) {
	if config.ContainerConfig != nil {
		return nm.deployContainerNode(config)
	}
	return nm.deployOnPremNode(config)
}

func (nm *NodeManager) deployContainerNode(config NodeConfig) (string, error) {
	containerConfig := &container.Config{
		Image: config.ContainerConfig.Image,
		Env:   config.ContainerConfig.Environment,
	}

	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{},
	}
	for hostPath, containerPath := range config.ContainerConfig.Volumes {
		hostConfig.Mounts = append(hostConfig.Mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: hostPath,
			Target: containerPath,
		})
	}

	resp, err := nm.client.ContainerCreate(nm.ctx, containerConfig, hostConfig, nil, nil, config.NodeID)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %v", err)
	}

	if err := nm.client.ContainerStart(nm.ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %v", err)
	}

	return resp.ID, nil
}

func (nm *NodeManager) deployOnPremNode(config NodeConfig) (string, error) {
	cmd := exec.Command("/usr/local/bin/blockchain-node", "--config", configFilePath(config))
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run node command: %v", err)
	}
	return out.String(), nil
}

func configFilePath(config NodeConfig) string {
	return fmt.Sprintf("/etc/blockchain/%s/%s.config", config.NetworkID, config.NodeID)
}

// VerifyNodeDeployment verifies the integrity of a node deployment
func (nm *NodeManager) VerifyNodeDeployment(nodeID string) error {
	containerInfo, err := nm.client.ContainerInspect(nm.ctx, nodeID)
	if err != nil {
		return fmt.Errorf("failed to inspect container: %v", err)
	}
	if !containerInfo.State.Running {
		return fmt.Errorf("container %s is not running", nodeID)
	}
	return nil
}

// GenerateNodeID generates a unique node ID
func GenerateNodeID() (string, error) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %v", err)
	}
	hash := sha256.Sum256(randomBytes)
	return hex.EncodeToString(hash[:]), nil
}

// EncryptData encrypts data using Argon2 and AES
func EncryptData(data string, password string) (string, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", fmt.Errorf("failed to generate salt: %v", err)
	}

	key := argon2.Key([]byte(password), salt, 3, 32*1024, 4, 32)

	encryptedData, err := aesEncrypt([]byte(data), key)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt data: %v", err)
	}

	combinedData := append(salt, encryptedData...)
	return hex.EncodeToString(combinedData), nil
}

// DecryptData decrypts data using Argon2 and AES
func DecryptData(encryptedDataHex string, password string) (string, error) {
	encryptedData, err := hex.DecodeString(encryptedDataHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted data: %v", err)
	}

	salt := encryptedData[:16]
	encrypted := encryptedData[16:]

	key := argon2.Key([]byte(password), salt, 3, 32*1024, 4, 32)

	decryptedData, err := aesDecrypt(encrypted, key)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt data: %v", err)
	}

	return string(decryptedData), nil
}

// aesEncrypt encrypts data using AES
func aesEncrypt(data []byte, key []byte) ([]byte, error) {
	// Implement AES encryption logic here
	// Dummy encryption logic for illustration
	return data, nil
}

// aesDecrypt decrypts data using AES
func aesDecrypt(data []byte, key []byte) ([]byte, error) {
	// Implement AES decryption logic here
	// Dummy decryption logic for illustration
	return data, nil
}

// MonitorNode monitors the status and performance of a node
func (nm *NodeManager) MonitorNode(nodeID string) (string, error) {
	containerInfo, err := nm.client.ContainerInspect(nm.ctx, nodeID)
	if err != nil {
		return "", fmt.Errorf("failed to inspect container: %v", err)
	}

	status := fmt.Sprintf("Node %s is running. Status: %s", nodeID, containerInfo.State.Status)
	return status, nil
}

// NodeHealthCheck performs a health check on a node
func (nm *NodeManager) NodeHealthCheck(nodeID string) error {
	resp, err := nm.client.ContainerStats(nm.ctx, nodeID, false)
	if err != nil {
		return fmt.Errorf("failed to get container stats: %v", err)
	}
	defer resp.Body.Close()

	// Parse the stats here and implement logic to check node health
	// Dummy health check for illustration
	return nil
}

// ScaleNodes scales the number of active nodes in the network
func (nm *NodeManager) ScaleNodes(targetNodeCount int, config NodeConfig) error {
	currentNodeCount, err := nm.getCurrentNodeCount(config.NetworkID)
	if err != nil {
		return fmt.Errorf("failed to get current node count: %v", err)
	}

	for i := currentNodeCount; i < targetNodeCount; i++ {
		nodeID, err := GenerateNodeID()
		if err != nil {
			return fmt.Errorf("failed to generate node ID: %v", err)
		}

		config.NodeID = nodeID
		_, err = nm.DeployNode(config)
		if err != nil {
			return fmt.Errorf("failed to deploy node: %v", err)
		}
	}

	return nil
}

func (nm *NodeManager) getCurrentNodeCount(networkID string) (int, error) {
	containers, err := nm.client.ContainerList(nm.ctx, types.ContainerListOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to list containers: %v", err)
	}

	nodeCount := 0
	for _, container := range containers {
		if container.Labels["networkID"] == networkID {
			nodeCount++
		}
	}

	return nodeCount, nil
}
