package pending_votes

import (
	"github.com/zarbchain/zarb-go/crypto"
	"github.com/zarbchain/zarb-go/errors"
	"github.com/zarbchain/zarb-go/proposal"
	"github.com/zarbchain/zarb-go/validator"
	"github.com/zarbchain/zarb-go/vote"
)

type PendingVotes struct {
	height     int
	validators []*validator.Validator
	roundVotes []*RoundVotes
}

func NewPendingVotes() *PendingVotes {
	pv := &PendingVotes{
		roundVotes: make([]*RoundVotes, 0),
	}
	return pv
}

func (pv *PendingVotes) GetRoundVotes(round int) *RoundVotes {
	if round < len(pv.roundVotes) {
		return pv.roundVotes[round]
	}
	return nil
}

func (pv *PendingVotes) HasVote(hash crypto.Hash) bool {
	for _, rv := range pv.roundVotes {
		if rv.HasVote(hash) {
			return true
		}
	}
	return false
}

func (pv *PendingVotes) MustGetRoundVotes(round int) *RoundVotes {
	for i := len(pv.roundVotes); i <= round; i++ {
		prepares := NewVoteSet(pv.height, i, vote.VoteTypePrepare, pv.validators)
		precommits := NewVoteSet(pv.height, i, vote.VoteTypePrecommit, pv.validators)
		rv := &RoundVotes{
			prepares:   prepares,
			precommits: precommits,
		}

		// extendind votes slice
		pv.roundVotes = append(pv.roundVotes, rv)
	}

	return pv.GetRoundVotes(round)
}

func (pv *PendingVotes) AddVote(v *vote.Vote) (bool, error) {
	if err := v.SanityCheck(); err != nil {
		return false, errors.Errorf(errors.ErrInvalidVote, "%v", err)
	}
	if v.Height() != pv.height {
		return false, errors.Errorf(errors.ErrInvalidVote, "Invalid height")
	}
	rv := pv.MustGetRoundVotes(v.Round())
	return rv.addVote(v)
}

func (pv *PendingVotes) PrepareVoteSet(round int) *VoteSet {
	rv := pv.MustGetRoundVotes(round)
	return rv.voteSet(vote.VoteTypePrepare)
}

func (pv *PendingVotes) PrecommitVoteSet(round int) *VoteSet {
	rv := pv.MustGetRoundVotes(round)
	return rv.voteSet(vote.VoteTypePrecommit)
}

func (pv *PendingVotes) HasRoundProposal(round int) bool {
	return pv.RoundProposal(round) != nil
}

func (pv *PendingVotes) RoundProposal(round int) *proposal.Proposal {
	rv := pv.GetRoundVotes(round)
	if rv == nil {
		return nil
	}
	return rv.proposal
}

func (pv *PendingVotes) SetRoundProposal(round int, proposal *proposal.Proposal) {
	rv := pv.MustGetRoundVotes(round)
	rv.proposal = proposal
}

func (pv *PendingVotes) MoveToNewHeight(height int, validators []*validator.Validator) {
	pv.height = height
	pv.roundVotes = make([]*RoundVotes, 0)
	pv.validators = validators
}

func (pv *PendingVotes) CanVote(addr crypto.Address) bool {
	for _, val := range pv.validators {
		if val.Address().EqualsTo(addr) {
			return true
		}
	}
	return false
}