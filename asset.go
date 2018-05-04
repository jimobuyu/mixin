package external

import (
	"crypto/md5"
	"fmt"
	"io"

	"github.com/satori/go.uuid"
)

const (
	BitcoinChainId         = "c6d0c728-2624-429b-8e0d-d9d19b6592fa"
	BitcoinCashChainId     = "fd11b6e3-0b87-41f1-a41f-f0e9b49e5bf0"
	LitecoinChainId        = "76c802a2-7c88-447f-a93e-c29c9e5dd9c8"
	EthereumChainId        = "43d61dcd-e413-450d-80b8-101d5e903357"
	EthereumClassicChainId = "2204c1ee-0ea2-4add-bb9a-b3719cfff93a"
	RippleChainId          = "23dfb5a5-5d7b-48b6-905f-3970e3176e27"
	SiacoinChainId         = "990c4c29-57e9-48f6-9819-7d986ea44985"
)

const (
	EthereumXINAssetKey   = "0xa974c709cfb4566686553a20790685a47aceaa33"
	EthereumEOSAssetKey   = "0x86fa049857e0209aa7d9e616f7eb3b3b78ecfdb0"
	EthereumBIGAssetKey   = "0x0396340f16bbec973280ab053efc3f208fa37795"
	EthereumPRSAssetKey   = "0xe0d95530820aafc51b1d98023aa1ff000b78d8b2"
	EthereumTRXAssetKey   = "0xf230b790e05390fc8295f4d3f60332c93bed42e2"
	EthereumOMGAssetKey   = "0xd26114cd6ee289accf82350c8d8487fedb8a0c07"
	EthereumBNBAssetKey   = "0xb8c77482e45f1f44de1745f52c74426c631bdd52"
	EthereumDGDAssetKey   = "0xe0b7927c4af23765cb51314a0e0521a9645f0e2a"
	EthereumSNTAssetKey   = "0x744d70fdbe2ba4cf95131626614a1763df805b9e"
	EthereumZRXAssetKey   = "0xe41d2489571d322189246dafa5ebde1f4699f498"
	EthereumBATAssetKey   = "0x0d8775f648430679a709e98d2b0cb6250d2887ef"
	EthereumKNCAssetKey   = "0xdd974d5c2e2928dea5f71b9825b8b646686bd200"
	EthereumDEWAssetKey   = "0x20e94867794dba030ee287f1406e100d03c84cd3"
	EthereumVENAssetKey   = "0xd850942ef8811f2a866692a623011bde52a462c1"
	EthereumCVCAssetKey   = "0x41e5560054824ea6b0732e656e3ad64e20e94e45"
	EthereumCNBAssetKey   = "0xec2a0550a2e4da2a027b3fc06f70ba15a94a6dac"
	EthereumICXAssetKey   = "0xb5a5f22694352c15b00323844ad545abb2b11028"
	EthereumCANDYAssetKey = "0xf2eab3a2034d3f6b63734d2e08262040e3ff7b48"
	EthereumIOSTAssetKey  = "0xfa1a856cfa3409cfa145fa4e20eb270df3eb21ab"
	EthereumRUFFAssetKey  = "0xf278c1ca969095ffddded020290cf8b5c424ace2"
	EthereumIHTAssetKey   = "0xeda8b016efa8b1161208cf041cd86972eee0f31e"
	EthereumDTAAssetKey   = "0x69b148395ce0015c13e36bffbad63f49ef874e03"
	EthereumTNBAssetKey   = "0xf7920b0768ecb20a123fac32311d07d193381d6f"
	EthereumLRCAssetKey   = "0xef68e7c694f40c8202821edf525de3782458639f"
	EthereumnCashAssetKey = "0x809826cceab68c387726af962713b64cb5cb3cca"
	EthereumnCREAssetKey  = "0x61f33da40594cec1e3dc900faf99f861d01e2e7d"
	EthereumnQUNAssetKey  = "0x264dc2dedcdcbb897561a57cba5085ca416fb7b4"
	EthereumnPAYAssetKey  = "0xb97048628db6b661d4c2aa833e95dbe1a905b280"
	EthereumnUSDTAssetKey = "0xdac17f958d2ee523a2206206994597c13d831ec7"
	EthereumAEAssetKey    = "0x5ca9a71b1d01849c0a95490cc00559717fcf0d1d"
	EthereumZILAssetKey   = "0x05f4a42e251f2d52b8ed15e9fedaacfcef1fad27"
	EthereumBTMAssetKey   = "0xcb97e65f07da24d46bcdd078ebebd7c6e6e3d750"
	EthereumPPTAssetKey   = "0xd4fa1460f537bb9085d22c7bccb5dd450ef28e3a"
	EthereumMKRAssetKey   = "0x9f8f72aa9304c8b593d555f12ef6589cc3a579a2"
	EthereumRHOCAssetKey  = "0x168296bb09e24a88805cb9c33356536b980d3fc5"
	EthereumGNTAssetKey   = "0xa74476443119a942de498590fe1f2454d7d4ac0d"
	EthereumWTCAssetKey   = "0xb7cb1c96db6b22b0d3d9536e0108d062bd488f74"
	EthereumAIONAssetKey  = "0x4ceda7906a5ed2179785cd3a40a69ee8bc99c466"
	EthereumREPAssetKey   = "0xe94327d07fc17907b4db788e5adf2ed424addff6"
	EthereumELFAssetKey   = "0xbf2179859fc6d5bee9bf9158632dc51678a4100e"
	EthereumLOOMAssetKey  = "0xa4e8c3ec456107ea67d3075bf9e3df3a75823db0"
	EthereumMITHAssetKey  = "0x3893b9422cd5d70a81edeffe3d5a1c6a978310bb"
	EthereumNASAssetKey   = "0x5d65d971895edc438f465c17db6992698a52318d"
	EthereumKCSAssetKey   = "0x039b5649a59967e3e936d7471f9c3700100ee1ab"
	EthereumBTOAssetKey   = "0x36905fc93280f52362a1cbab151f25dc46742fb5"
	EthereumNULSAssetKey  = "0xb91318f35bdb262e9423bc7c7c2a3a93dd93c92c"
)

func GetFeeAsset(chainId string) (string, error) {
	switch chainId {
	case RippleChainId:
		return RippleChainId, nil
	case SiacoinChainId:
		return SiacoinChainId, nil
	case EthereumChainId:
		return EthereumChainId, nil
	case EthereumClassicChainId:
		return EthereumClassicChainId, nil
	case BitcoinChainId:
		return BitcoinChainId, nil
	case BitcoinCashChainId:
		return BitcoinCashChainId, nil
	case LitecoinChainId:
		return LitecoinChainId, nil
	}
	return "", fmt.Errorf("unsupported chain id %s", chainId)
}

func UniqueAssetId(chainId, assetAddress string) string {
	h := md5.New()
	io.WriteString(h, chainId)
	io.WriteString(h, assetAddress)
	sum := h.Sum(nil)
	sum[6] = (sum[6] & 0x0f) | 0x30
	sum[8] = (sum[8] & 0x3f) | 0x80
	return uuid.FromBytesOrNil(sum).String()
}

func BaseBlock(chainId string) int64 {
	switch chainId {
	case RippleChainId:
		return 37944664
	case SiacoinChainId:
		return 152927
	case EthereumChainId:
		return 5481185
	case EthereumClassicChainId:
		return 5747975
	case BitcoinChainId:
		return 519288
	case BitcoinCashChainId:
		return 526881
	case LitecoinChainId:
		return 1406506
	}
	return 0
}
