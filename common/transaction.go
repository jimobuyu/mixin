package common

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/MixinNetwork/mixin/config"
	"github.com/MixinNetwork/mixin/crypto"
)

const (
	TxVersion      = 0x01
	ExtraSizeLimit = 256

	OutputTypeScript       = 0x00
	OutputTypeWithdrawal   = 0xa1
	OutputTypeSlash        = 0xa2
	OutputTypeNodePledge   = 0xa3
	OutputTypeNodeAccept   = 0xa4
	OutputTypeNodeDepart   = 0xa5
	OutputTypeNodeRemove   = 0xa6
	OutputTypeDomainAccept = 0xa7
	OutputTypeDomainRemove = 0xa8
)

type Input struct {
	Hash    crypto.Hash  `json:"hash,omitempty"`
	Index   int          `json:"index,omitempty"`
	Genesis []byte       `json:"genesis,omitempty"`
	Deposit *DepositData `json:"deposit,omitempty"`
	Rebate  []byte       `json:"rebate,omitempty"`
	Mint    []byte       `json:"mint,omitempty"`
}

type Output struct {
	Type   uint8        `json:"type"`
	Amount Integer      `json:"amount"`
	Keys   []crypto.Key `json:"keys,omitempty"`

	// OutputTypeScript fields
	Script Script     `json:"script,omitempty"`
	Mask   crypto.Key `json:"mask,omitempty"`
}

type DepositData struct {
	Chain           crypto.Hash `json:"chain"`
	AssetKey        string      `json:"asset"`
	TransactionHash string      `json:"transaction"`
	Amount          Integer     `json:"amount"`
}

type Transaction struct {
	Version uint8       `json:"version"`
	Asset   crypto.Hash `json:"asset"`
	Inputs  []*Input    `json:"inputs"`
	Outputs []*Output   `json:"outputs"`
	Extra   []byte      `json:"extra,omitempty"`
	Hash    crypto.Hash `msgpack:"-"json:"hash"`
}

type SignedTransaction struct {
	Transaction
	Signatures [][]crypto.Signature `json:"signatures,omitempty"`
}

func (tx *Transaction) ViewGhostKey(a *crypto.Key) []*Output {
	outputs := make([]*Output, 0)

	for i, o := range tx.Outputs {
		if o.Type != OutputTypeScript {
			continue
		}

		out := &Output{
			Type:   o.Type,
			Amount: o.Amount,
			Script: o.Script,
			Mask:   o.Mask,
		}
		for _, k := range o.Keys {
			key := crypto.ViewGhostOutputKey(&k, a, &o.Mask, uint64(i))
			out.Keys = append(out.Keys, *key)
		}
		outputs = append(outputs, out)
	}

	return outputs
}

func (tx *SignedTransaction) Validate(store DataStore) error {
	if tx.Version != TxVersion {
		return fmt.Errorf("invalid tx version %d", tx.Version)
	}

	if len(tx.Inputs) != len(tx.Signatures) {
		return fmt.Errorf("invalid tx signature number %d %d", len(tx.Inputs), len(tx.Signatures))
	}

	if len(tx.Extra) > ExtraSizeLimit {
		return fmt.Errorf("invalid extra size %d", len(tx.Extra))
	}

	msg := MsgpackMarshalPanic(tx.Transaction)
	if len(msg) > config.TransactionMaximumSize {
		return fmt.Errorf("invalid transaction size %d", len(msg))
	}

	var inputAmount, outputAmount Integer

	inputsFilter := make(map[string]*UTXO)
	for i, in := range tx.Inputs {
		if len(in.Genesis) > 0 {
			return fmt.Errorf("invalid genesis input detected %s", hex.EncodeToString(in.Genesis))
		}
		if in.Deposit != nil {
			err := tx.validateDepositInput(store, msg)
			if err != nil {
				return err
			}
			err = store.CheckDepositInput(in.Deposit, tx.PayloadHash())
			if err != nil {
				return err
			}
			inputAmount = in.Deposit.Amount
			break
		}

		fk := fmt.Sprintf("%s:%d", in.Hash.String(), in.Index)
		if inputsFilter[fk] != nil {
			return fmt.Errorf("invalid input %s", fk)
		}

		utxo, err := store.ReadUTXO(in.Hash, in.Index)
		if err != nil {
			return err
		}
		if utxo == nil {
			return fmt.Errorf("input not found %s:%d", in.Hash.String(), in.Index)
		}
		if utxo.Asset.String() != tx.Asset.String() {
			return fmt.Errorf("invalid input asset %s %s", utxo.Asset.String(), tx.Asset.String())
		}

		err = validateUTXO(utxo, tx.Signatures[i], msg)
		if err != nil {
			return err
		}
		inputsFilter[fk] = utxo
		inputAmount = inputAmount.Add(utxo.Amount)
	}

	outputsFilter := make(map[crypto.Key]bool)
	for _, o := range tx.Outputs {
		if o.Amount.Sign() <= 0 {
			return fmt.Errorf("invalid output amount %s", o.Amount.String())
		}
		for _, k := range o.Keys {
			if outputsFilter[k] {
				return fmt.Errorf("invalid output key %s", k.String())
			}
			outputsFilter[k] = true
			exist, err := store.CheckGhost(k)
			if err != nil {
				return err
			} else if exist {
				return fmt.Errorf("invalid output key %s", k.String())
			}
		}
		outputAmount = outputAmount.Add(o.Amount)

		switch o.Type {
		case OutputTypeScript:
			for _, in := range inputsFilter {
				if in.Type != OutputTypeScript {
					return fmt.Errorf("invalid utxo type %d", in.Type)
				}
			}
			err := o.Script.VerifyFormat()
			if err != nil {
				return err
			}
		case OutputTypeNodePledge:
			for _, in := range inputsFilter {
				if in.Type != OutputTypeScript {
					return fmt.Errorf("invalid utxo type %d", in.Type)
				}
			}
			err := tx.validateNodePledge(store)
			if err != nil {
				return err
			}
		case OutputTypeNodeAccept:
			for _, in := range inputsFilter {
				if in.Type != OutputTypeNodePledge && in.Type != OutputTypeNodeAccept {
					return fmt.Errorf("invalid utxo type %d", in.Type)
				}
			}
			err := tx.validateNodeAccept(store, inputAmount)
			if err != nil {
				return err
			}
		}
	}

	if inputAmount.Cmp(outputAmount) != 0 {
		return fmt.Errorf("invalid input output amount %s %s", inputAmount.String(), outputAmount.String())
	}
	return nil
}

func (tx *SignedTransaction) validateDepositInput(store DataStore, msg []byte) error {
	if len(tx.Inputs) != 1 {
		return fmt.Errorf("invalid inputs count %d for deposit", len(tx.Inputs))
	}
	if len(tx.Signatures) != 1 || len(tx.Signatures[0]) != 1 {
		return fmt.Errorf("invalid signatures count %d for deposit", len(tx.Signatures))
	}
	sig, valid := tx.Signatures[0][0], false
	domains := store.ReadDomains()
	for _, d := range domains {
		if d.Account.PublicSpendKey.Verify(msg, sig) {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid domain signature for deposit")
	}
	return nil
}

func (tx *Transaction) validateNodePledge(store DataStore) error {
	if len(tx.Outputs) != 1 {
		return fmt.Errorf("invalid outputs count %d for pledge transaction", len(tx.Outputs))
	}
	o := tx.Outputs[0]
	if o.Amount.Cmp(NewInteger(10000)) != 0 {
		return fmt.Errorf("invalid pledge amount %s", o.Amount.String())
	}
	nodes := store.ReadConsensusNodes()
	for _, n := range nodes {
		if n.State != NodeStateAccepted {
			return fmt.Errorf("invalid node pending state %s %s", n.Account.String(), n.State)
		}
	}

	var publicSpend crypto.Key
	copy(publicSpend[:], tx.Extra)
	privateView := publicSpend.DeterministicHashDerive()
	acc := Address{
		PrivateViewKey: privateView,
		PublicViewKey:  privateView.Public(),
		PublicSpendKey: publicSpend,
	}
	nodes = append(nodes, Node{Account: acc})
	if len(nodes) != len(o.Keys) {
		return fmt.Errorf("invalid output keys count %d %d for pledge transaction", len(nodes), len(o.Keys))
	}

	if o.Script.VerifyFormat() != nil || int(o.Script[2]) != len(nodes)*2/3+1 {
		return fmt.Errorf("invalid output script %s %d", o.Script, len(nodes)*2/3+1)
	}

	filter := make(map[crypto.Key]bool)
	for _, n := range nodes {
		filter[n.Account.PublicSpendKey] = true
	}
	for i, k := range o.Keys {
		for _, n := range nodes {
			ghost := crypto.ViewGhostOutputKey(&k, &n.Account.PrivateViewKey, &o.Mask, 0)
			delete(filter, *ghost)
		}
		if len(filter) != len(nodes)-1-i {
			return fmt.Errorf("invalid output keys signatures %d", len(filter))
		}
	}
	return nil
}

func (tx *Transaction) validateNodeAccept(store DataStore, inputAmount Integer) error {
	if len(tx.Outputs) != 1 {
		return fmt.Errorf("invalid outputs count %d for accept transaction", len(tx.Outputs))
	}
	if len(tx.Inputs) != 2 {
		return fmt.Errorf("invalid inputs count %d for accept transaction", len(tx.Inputs))
	}
	var pledging *Node
	filter := make(map[string]string)
	nodes := store.ReadConsensusNodes()
	for _, n := range nodes {
		filter[n.Account.String()] = n.State
		if n.State == NodeStateDeparting {
			return fmt.Errorf("invalid node pending state %s %s", n.Account.String(), n.State)
		}
		if n.State == NodeStateAccepted {
			continue
		}
		if n.State == NodeStatePledging && pledging == nil {
			pledging = &n
		} else {
			return fmt.Errorf("invalid pledging nodes %s %s", pledging.Account.String(), n.Account.String())
		}
	}
	if pledging == nil {
		return fmt.Errorf("no pledging node needs to get accepted")
	}
	nodesAmount := NewInteger(uint64(10000 * len(nodes)))
	if inputAmount.Cmp(nodesAmount) != 0 {
		return fmt.Errorf("invalid accept input amount %s %s", inputAmount.String(), nodesAmount.String())
	}

	lastAccept, err := store.ReadTransaction(tx.Inputs[0].Hash)
	if err != nil {
		return err
	}
	ao := lastAccept.Outputs[0]
	if len(lastAccept.Outputs) != 1 {
		return fmt.Errorf("invalid accept utxo count %d", len(lastAccept.Outputs))
	}
	if ao.Type != OutputTypeNodeAccept {
		return fmt.Errorf("invalid accept utxo type %d", ao.Type)
	}
	var publicSpend crypto.Key
	copy(publicSpend[:], lastAccept.Extra)
	privateView := publicSpend.DeterministicHashDerive()
	acc := Address{
		PublicViewKey:  privateView.Public(),
		PublicSpendKey: publicSpend,
	}
	if filter[acc.String()] != NodeStateAccepted {
		return fmt.Errorf("invalid accept utxo source %s", filter[acc.String()])
	}

	lastPledge, err := store.ReadTransaction(tx.Inputs[1].Hash)
	if err != nil {
		return err
	}
	po := lastPledge.Outputs[0]
	if len(lastPledge.Outputs) != 1 {
		return fmt.Errorf("invalid pledge utxo count %d", len(lastPledge.Outputs))
	}
	if po.Type != OutputTypeNodePledge {
		return fmt.Errorf("invalid pledge utxo type %d", po.Type)
	}
	copy(publicSpend[:], lastPledge.Extra)
	privateView = publicSpend.DeterministicHashDerive()
	acc = Address{
		PublicViewKey:  privateView.Public(),
		PublicSpendKey: publicSpend,
	}
	if filter[acc.String()] != NodeStatePledging {
		return fmt.Errorf("invalid pledge utxo source %s", filter[acc.String()])
	}
	return nil
}

func validateUTXO(utxo *UTXO, sigs []crypto.Signature, msg []byte) error {
	switch utxo.Type {
	case OutputTypeScript:
	case OutputTypeNodePledge:
	case OutputTypeNodeAccept:
	default:
		return fmt.Errorf("invalid input type %d", utxo.Type)
	}

	var offset, valid int
	for _, sig := range sigs {
		for i, k := range utxo.Keys {
			if i < offset {
				continue
			}
			if k.Verify(msg, sig) {
				valid = valid + 1
				offset = i + 1
			}
		}
	}

	return utxo.Script.Validate(valid)
}

func (tx *Transaction) PayloadHash() crypto.Hash {
	msg := MsgpackMarshalPanic(tx)
	return crypto.NewHash(msg)
}

func (tx *SignedTransaction) Marshal() []byte {
	return MsgpackMarshalPanic(tx)
}

func (signed *SignedTransaction) SignInput(reader UTXOReader, index int, accounts []Address) error {
	msg := MsgpackMarshalPanic(signed.Transaction)

	if len(accounts) == 0 {
		return nil
	}
	if index >= len(signed.Inputs) {
		return fmt.Errorf("invalid input index %d/%d", index, len(signed.Inputs))
	}
	in := signed.Inputs[index]
	if in.Deposit != nil || len(in.Rebate) > 0 || len(in.Mint) > 0 {
		return signed.SignRaw(accounts[0].PrivateSpendKey)
	}

	utxo, err := reader.ReadUTXO(in.Hash, in.Index)
	if err != nil {
		return err
	}
	if utxo == nil {
		return fmt.Errorf("input not found %s:%d", in.Hash.String(), in.Index)
	}

	keysFilter := make(map[string]bool)
	for _, k := range utxo.Keys {
		keysFilter[k.String()] = true
	}

	sigs := make([]crypto.Signature, 0)
	for _, acc := range accounts {
		priv := crypto.DeriveGhostPrivateKey(&utxo.Mask, &acc.PrivateViewKey, &acc.PrivateSpendKey, uint64(in.Index))
		if keysFilter[priv.Public().String()] {
			sigs = append(sigs, priv.Sign(msg))
		}
	}
	signed.Signatures = append(signed.Signatures, sigs)
	return nil
}

func (signed *SignedTransaction) SignRaw(key crypto.Key) error {
	msg := MsgpackMarshalPanic(signed.Transaction)

	if len(signed.Inputs) != 1 {
		return fmt.Errorf("invalid inputs count %d", len(signed.Inputs))
	}
	in := signed.Inputs[0]
	if in.Deposit == nil && len(in.Rebate) == 0 && len(in.Mint) == 0 {
		return fmt.Errorf("invalid input format")
	}
	signed.Signatures = append(signed.Signatures, []crypto.Signature{key.Sign(msg)})
	return nil
}

func NewTransaction(asset crypto.Hash) *Transaction {
	return &Transaction{
		Version: TxVersion,
		Asset:   asset,
	}
}

func (tx *Transaction) AddDepositInput(data *DepositData) {
	tx.Inputs = append(tx.Inputs, &Input{
		Deposit: data,
	})
}

func (tx *Transaction) AddInput(hash crypto.Hash, index int) {
	in := &Input{
		Hash:  hash,
		Index: index,
	}
	tx.Inputs = append(tx.Inputs, in)
}

func (tx *Transaction) AddScriptOutput(accounts []Address, s Script, amount Integer) error {
	seed := make([]byte, 64)
	_, err := rand.Read(seed)
	if err != nil {
		return err
	}
	r := crypto.NewKeyFromSeed(seed)
	R := r.Public()
	out := &Output{
		Type:   OutputTypeScript,
		Amount: amount,
		Script: s,
		Mask:   R,
		Keys:   make([]crypto.Key, 0),
	}

	for _, a := range accounts {
		k := crypto.DeriveGhostPublicKey(&r, &a.PublicViewKey, &a.PublicSpendKey, uint64(len(tx.Outputs)))
		out.Keys = append(out.Keys, *k)
	}
	tx.Outputs = append(tx.Outputs, out)
	return nil
}
