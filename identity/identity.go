package identity

import (
	"github.com/libp2p/go-libp2p/core/crypto"
	"go.uber.org/zap"
	"os"
)

type Identity struct {
	log *zap.Logger
}

func New(log *zap.Logger) *Identity {
	return &Identity{
		log: log.Named("identity"),
	}
}

// generate a private key and store it at the given path
func (i *Identity) GeneratePrivateKey(path string) (crypto.PrivKey, error) {
	privateKey, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 0)

	if err != nil {
		i.log.Error("failed to generate key pair", zap.Error(err))
		return nil, err
	}

	bytes, err := crypto.MarshalPrivateKey(privateKey)

	if err != nil {
		i.log.Error("failed to marshal private key", zap.Error(err))
		return nil, err
	}

	if err := os.WriteFile(path, bytes, 0400); err != nil {
		i.log.Error("failed to write identity file",
			zap.String("path", path),
			zap.Error(err),
		)
		return nil, err
	}

	return privateKey, err
}

// attempt to read private key from a given path, if not availabe then generate it at the given path
func (i *Identity) LoadIdentity(keyPath string) (crypto.PrivKey, error) {
	if _, err := os.Stat(keyPath); err == nil {

		i.log.Info("reading identity", zap.String("path", keyPath))
		return i.ReadIdentity(keyPath)

	} else if os.IsNotExist(err) {

		i.log.Info("generating identity", zap.String("path", keyPath))
		return i.GeneratePrivateKey(keyPath)

	} else {

		i.log.Error("failed to stat identity file",
			zap.String("path", keyPath),
			zap.Error(err),
		)
		return nil, err

	}
}

// read the private key from the given path
func (i *Identity) ReadIdentity(path string) (crypto.PrivKey, error) {
	bytes, err := os.ReadFile(path)

	if err != nil {
		i.log.Error("failed to read identity file",
			zap.String("path", path),
			zap.Error(err),
		)
		return nil, err
	}

	return crypto.UnmarshalPrivateKey(bytes)
}
