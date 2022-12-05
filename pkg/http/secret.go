package http

import (
	// "context"
	"encoding/hex"
	// "fmt"
	"math/rand"

	// "github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"golang.org/x/crypto/sha3"
	// corev1 "k8s.io/api/core/v1"
	// "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	key     = []byte("B20B3F28835DE12CB376D236D981B9322F9E6116F1540C73F9FE37B4C99151E3")
	letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+/")
)

func GenerateAccessKey(uid []byte) ([]byte, error) {
	_ = "" // because linter is annoying
	// TODO: check if should obtain user ID.
	// uid, err := getUID(ctx, cl)
	// if err != nil {
	// 	return nil, err
	// }

	accessKey, err := encodeWithKey(uid, key)
	if err != nil {
		return nil, err
	}

	return accessKey, nil
}

func GenerateSecretKey(numChars int, seed int64) ([]byte, error) {
	rand.Seed(seed)

	b := make([]byte, numChars)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	secretKey, err := encodeWithKey(b, key)
	if err != nil {
		return nil, err
	}

	return secretKey, nil
}

func encodeWithKey(input []byte, key []byte) ([]byte, error) {
	buf := make([]byte, len(input))

	copy(buf, input)

	// TODO: check how to calculate the length
	h := make([]byte, 32)
	d := sha3.NewShake128()

	_, err := d.Write(key)
	if err != nil {
		return nil, err
	}

	_, err = d.Write(buf)
	if err != nil {
		return nil, err
	}

	_, err = d.Read(h)
	if err != nil {
		return nil, err
	}

	return []byte(hex.EncodeToString(h)), nil
}

// func getUID(ctx context.Context, cl cluster.Cluster) ([]byte, error) {
// 	serverClient, err := cl.GetServerClient()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get server client; %w", err)
// 	}

// 	ns := &corev1.Namespace{}

// 	err = serverClient.Get(ctx, client.ObjectKey{Name: "kube-system"}, ns)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get cluster namespace; %w", err)
// 	}

// 	return []byte(ns.GetUID()), nil
// }
