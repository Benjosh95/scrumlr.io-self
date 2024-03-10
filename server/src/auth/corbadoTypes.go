package auth

import (
	"github.com/corbado/corbado-go"
	"github.com/corbado/corbado-go/pkg/stdlib"
)

type CombinedSDK struct {
	CorbadoSDK       *corbado.Impl
	CorbadoSDKHelper *stdlib.SDKHelpers
}
