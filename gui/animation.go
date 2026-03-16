// Package gui contains the Ebiten-based GUI for Ashta Changa.
package gui

// AnimationManager handles all game animations
type AnimationManager struct {
	// Roll animation counter (0 = idle)
	animRoll int

	// Token movement animation
	animMove *AnimMoveState

	// Kill animation
	animKill *AnimKillState

	// Extra turn flash counter
	animExtra int

	// Last drawn token screen positions [player][token][x,y]
	lastTokenPos [4][4][2]float64
}

// AnimMoveState represents a token movement animation
type AnimMoveState struct {
	PlayerIdx int
	TokenIdx  int
	FromX     float64
	FromY     float64
	ToX       float64
	ToY       float64
	Progress  int // 0-30 frames
}

// AnimKillState represents a kill animation
type AnimKillState struct {
	PlayerIdx int
	TokenIdx  int
	HomeX     float64
	HomeY     float64
	Progress  int // 0-30 frames
}

// NewAnimationManager creates a new animation manager
func NewAnimationManager() *AnimationManager {
	return &AnimationManager{}
}

// Tick updates all active animations
func (am *AnimationManager) Tick() {
	// Roll animation: spins for 30 frames then stops
	if am.animRoll > 0 {
		am.animRoll--
	}

	// Move animation: progresses from 0 to 30
	if am.animMove != nil {
		am.animMove.Progress++
		if am.animMove.Progress >= 30 {
			am.animMove = nil
		}
	}

	// Kill animation: progresses from 0 to 30
	if am.animKill != nil {
		am.animKill.Progress++
		if am.animKill.Progress >= 30 {
			am.animKill = nil
		}
	}

	// Extra turn flash
	if am.animExtra > 0 {
		am.animExtra--
	}
}

// StartRollAnim starts the pre-roll shell animation
func (am *AnimationManager) StartRollAnim() {
	am.animRoll = 30 // ~0.5 seconds at 60fps
}

// StartMoveAnim starts token movement animation
func (am *AnimationManager) StartMoveAnim(playerIdx, tokenIdx int, fromX, fromY, toX, toY float64) {
	am.animMove = &AnimMoveState{
		PlayerIdx: playerIdx,
		TokenIdx:  tokenIdx,
		FromX:     fromX,
		FromY:     fromY,
		ToX:       toX,
		ToY:       toY,
		Progress:  0,
	}
}

// StartKillAnim starts kill animation (token returning home)
func (am *AnimationManager) StartKillAnim(playerIdx, tokenIdx int, homeX, homeY float64) {
	am.animKill = &AnimKillState{
		PlayerIdx: playerIdx,
		TokenIdx:  tokenIdx,
		HomeX:     homeX,
		HomeY:     homeY,
		Progress:  0,
	}
}

// StartExtraAnim starts extra turn flash
func (am *AnimationManager) StartExtraAnim() {
	am.animExtra = 60 // ~1 second flash
}

// IsRolling returns true if roll animation is active
func (am *AnimationManager) IsRolling() bool {
	return am.animRoll > 0
}

// IsMoving returns true if move animation is active
func (am *AnimationManager) IsMoving() bool {
	return am.animMove != nil
}

// IsKilling returns true if kill animation is active
func (am *AnimationManager) IsKilling() bool {
	return am.animKill != nil
}

// IsExtraTurnFlashing returns true if extra turn flash is active
func (am *AnimationManager) IsExtraTurnFlashing() bool {
	return am.animExtra > 0
}

// GetRollAnimCount returns the current roll animation counter
func (am *AnimationManager) GetRollAnimCount() int {
	return am.animRoll
}

// GetMoveAnim returns the current move animation state
func (am *AnimationManager) GetMoveAnim() *AnimMoveState {
	return am.animMove
}

// GetKillAnim returns the current kill animation state
func (am *AnimationManager) GetKillAnim() *AnimKillState {
	return am.animKill
}

// GetExtraAnimCount returns the current extra turn animation counter
func (am *AnimationManager) GetExtraAnimCount() int {
	return am.animExtra
}

// GetLastTokenPos returns the last drawn position for a token
func (am *AnimationManager) GetLastTokenPos(playerIdx, tokenIdx int) (float64, float64) {
	return am.lastTokenPos[playerIdx][tokenIdx][0], am.lastTokenPos[playerIdx][tokenIdx][1]
}

// SetLastTokenPos sets the last drawn position for a token
func (am *AnimationManager) SetLastTokenPos(playerIdx, tokenIdx int, x, y float64) {
	am.lastTokenPos[playerIdx][tokenIdx][0] = x
	am.lastTokenPos[playerIdx][tokenIdx][1] = y
}
