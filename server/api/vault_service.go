package api

import (

	"github.com/cmrd-a/GophKeeper/gen/proto/v1/vault"

)

// VaultServer implements VaultService.
type VaultServer struct {
	vault.UnimplementedVaultServiceServer
}
