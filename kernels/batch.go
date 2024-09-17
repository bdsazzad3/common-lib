package kernels

import (
	"math/big"

	"github.com/pkg/errors"

	"github.com/cornbase/common-lib/validate"
)

// EarnRequestFullBatch is a batch of unrelated earn requests
type EarnRequestFullBatch struct {
	UserAddrs   []string `json:"userAddrs"`
	Sources     []string `json:"sources"`
	SubSources  []string `json:"subSources"`
	SourceUsers []string `json:"-"`
	StartBlocks []int64  `json:"startBlocks"`
	StartTimes  []int64  `json:"startTimes"`
	EarnRates   []string `json:"earnRates"`
}

func (e EarnRequestFullBatch) IsPerBlock() bool {
	return len(e.StartBlocks) != 0 // If startBlock is set, then it's per block
}

func (e EarnRequestFullBatch) Size() int {
	return len(e.UserAddrs)
}

func (e EarnRequestFullBatch) Validate() error {
	if len(e.UserAddrs) == 0 {
		return ErrEmptyBatch
	}
	if len(e.UserAddrs) != len(e.Sources) {
		return errors.New("userAddrs and sources must be the same length")
	}
	if len(e.UserAddrs) != len(e.SubSources) {
		return errors.New("userAddrs and subSources must be the same length")
	}
	if len(e.UserAddrs) != len(e.StartBlocks) {
		return errors.New("userAddrs and startBlocks must be the same length")
	}
	if len(e.UserAddrs) != len(e.StartTimes) {
		return errors.New("userAddrs and startTimes must be the same length")
	}
	if len(e.StartBlocks) != 0 && len(e.UserAddrs) != len(e.StartBlocks) {
		return errors.New("startBlocks must be the same length as userAddrs or empty/null")
	}
	if len(e.UserAddrs) != len(e.EarnRates) {
		return errors.New("userAddrs and earnRates must be the same length")
	}
	if len(e.SourceUsers) != 0 {
		return errors.New("sourceUsers must be empty")
	}

	for _, userAddr := range e.UserAddrs {
		_, err := validate.GetValidEthAddr(userAddr)
		if err != nil {
			return err
		}
	}

	for _, source := range e.Sources {
		if len(source) == 0 {
			return ErrEmptySource
		}
	}

	for _, subSource := range e.SubSources {
		if len(subSource) == 0 {
			return ErrEmptySubSource
		}
	}

	for _, startBlock := range e.StartBlocks {
		if startBlock < 0 {
			return ErrNonPostiveStartBlock
		}
	}

	for _, startTime := range e.StartTimes {
		if startTime < 1 {
			return ErrNonPostiveStartTime
		}
	}

	for _, earnRate := range e.EarnRates {
		floatVal, ok := big.NewFloat(0).SetString(earnRate)
		if !ok {
			return ErrInvalidEarnRate
		}

		if floatVal.Sign() < 0 {
			return ErrNegativeRate
		}

		if floatVal.IsInf() {
			return ErrEarnInf
		}
	}

	return nil
}

func BatchUnrelatedEarnRequests(earnRequests []EarnRequest) (EarnRequestFullBatch, error) {
	if len(earnRequests) == 0 {
		return EarnRequestFullBatch{}, ErrEmptyBatch
	}

	out := EarnRequestFullBatch{
		UserAddrs:   make([]string, len(earnRequests)),
		Sources:     make([]string, len(earnRequests)),
		SubSources:  make([]string, len(earnRequests)),
		SourceUsers: make([]string, len(earnRequests)),
		StartBlocks: nil,
		StartTimes:  make([]int64, len(earnRequests)),
		EarnRates:   make([]string, len(earnRequests)),
	}

	if earnRequests[0].StartBlock != 0 {
		out.StartBlocks = make([]int64, len(earnRequests))
	}

	for i := range earnRequests {
		out.UserAddrs[i] = earnRequests[i].UserAddr
		out.EarnRates[i] = earnRequests[i].EarnRate
		out.SourceUsers[i] = earnRequests[i].GetSourceUser()
		out.Sources[i] = earnRequests[i].Source
		out.SubSources[i] = earnRequests[i].SubSource
		out.StartTimes[i] = earnRequests[i].StartTime
		if out.StartBlocks != nil {
			out.StartBlocks[i] = earnRequests[i].StartBlock
		}
	}

	return out, nil
}

// EarnRequestBatch is a batch of related earn requests
type EarnRequestBatch struct {
	UserAddrs   []string `json:"userAddrs"`
	Source      string   `json:"source"`
	SubSource   string   `json:"subSource"`
	SourceUsers []string `json:"-"`
	StartBlock  int64    `json:"startBlock"`
	StartTime   int64    `json:"startTime"`
	EarnRates   []string `json:"earnRates"`
}

func (e EarnRequestBatch) IsPerBlock() bool {
	return e.StartBlock != 0 // If startBlock is set, then it's per block
}

func (e EarnRequestBatch) Size() int {
	return len(e.UserAddrs)
}

func (e EarnRequestBatch) Validate() error {
	if e.StartTime == 0 { // Start time is always required
		return ErrMissingStart
	}

	if len(e.UserAddrs) != len(e.EarnRates) {
		return errors.Errorf("userAddrs and earnRates must be the same length")
	}

	for _, userAddr := range e.UserAddrs {
		_, err := validate.GetValidEthAddr(userAddr)
		if err != nil {
			return err
		}
	}

	for _, earnRate := range e.EarnRates {
		floatVal, ok := big.NewFloat(0).SetString(earnRate)
		if !ok {
			return ErrInvalidEarnRate
		}

		if floatVal.Sign() < 0 {
			return ErrNegativeRate
		}

		if floatVal.IsInf() {
			return ErrEarnInf
		}
	}

	if e.StartBlock < 0 {
		return ErrNonPostiveStartBlock
	}

	if e.StartTime < 1 {
		return ErrNonPostiveStartTime
	}

	if len(e.Source) == 0 {
		return ErrEmptySource
	}

	if len(e.SubSource) == 0 {
		return ErrEmptySubSource
	}

	return nil
}

func (e EarnRequestBatch) WithReferralBonuses(referralChains [][]string, tierEarnRates map[int]float64) (EarnRequestBatch, error) {

	out := EarnRequestBatch{
		UserAddrs:   make([]string, len(e.UserAddrs)),
		Source:      e.Source,
		SubSource:   e.SubSource,
		SourceUsers: make([]string, len(e.UserAddrs)),
		StartBlock:  e.StartBlock,
		StartTime:   e.StartTime,
		EarnRates:   make([]string, len(e.EarnRates)),
	}

	copy(out.UserAddrs, e.UserAddrs)
	copy(out.EarnRates, e.EarnRates)
	copy(out.SourceUsers, e.UserAddrs)

	for i := range referralChains {
		earnRate, ok := big.NewFloat(0).SetString(e.EarnRates[i])
		if !ok {
			return EarnRequestBatch{}, errors.New("invalid earn rate")
		}
		for j := range referralChains[i] {
			earnRateTier := big.NewFloat(tierEarnRates[j])
			earnRateTier.Mul(earnRate, earnRateTier)
			out.UserAddrs = append(out.UserAddrs, referralChains[i][j])
			out.SourceUsers = append(out.SourceUsers, e.UserAddrs[i])
			out.EarnRates = append(out.EarnRates, earnRateTier.String())
		}
	}

	return out, nil
}

func BatchEarnRequests(earnRequests []EarnRequest) (EarnRequestBatch, error) {
	if len(earnRequests) == 0 {
		return EarnRequestBatch{}, ErrEmptyBatch
	}
	out := EarnRequestBatch{
		UserAddrs:   make([]string, len(earnRequests)),
		Source:      earnRequests[0].Source,
		SubSource:   earnRequests[0].SubSource,
		SourceUsers: make([]string, len(earnRequests)),
		StartBlock:  earnRequests[0].StartBlock,
		StartTime:   earnRequests[0].StartTime,
		EarnRates:   make([]string, len(earnRequests)),
	}

	for i := range earnRequests {
		out.UserAddrs[i] = earnRequests[i].UserAddr
		out.EarnRates[i] = earnRequests[i].EarnRate
		out.SourceUsers[i] = earnRequests[i].GetSourceUser()
		if out.StartBlock != earnRequests[i].StartBlock {
			return EarnRequestBatch{}, errors.New("startBlock must be the same for all requests")
		}
		if out.StartTime != earnRequests[i].StartTime {
			return EarnRequestBatch{}, errors.New("startTime must be the same for all requests")
		}
		if out.Source != earnRequests[i].Source {
			return EarnRequestBatch{}, errors.New("source must be the same for all requests")
		}
		if out.SubSource != earnRequests[i].SubSource {
			return EarnRequestBatch{}, errors.New("subSource must be the same for all requests")
		}
	}

	return out, nil
}