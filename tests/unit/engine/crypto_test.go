// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/engine/crypto_test.go
// @id            test.unit.engine.crypto
// @layer         Layer 4: Tier-2 Unit Verification Suite (Engine Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.2.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @system        FaithNet OSI L5/L6 Sovereign Crypto & .gen/.rev Dual-Witness Suite
// @role          Strict 1:1 Unit Isolation Gate for cog.engine.crypto
// ==============================================================================

package engine_unit_test

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ==============================================================================

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================

func TestCryptoUnit_Sec1_NovaZivNodeIDAuthAndGenRevRoundtrip(t *testing.T) {
	nova, err := engine.GenerateSovereignIdentity()
	if err != nil {
		t.Fatalf("[SEC 1] GenerateSovereignIdentity Nova failed: %v", err)
	}
	ziv, err := engine.GenerateSovereignIdentity()
	if err != nil {
		t.Fatalf("[SEC 1] GenerateSovereignIdentity Ziv failed: %v", err)
	}
	if !nova.NodeID.IsValid() || !ziv.NodeID.IsValid() || nova.NodeID == ziv.NodeID {
		t.Fatalf("[SEC 1] invalid or colliding NodeIDs: %s vs %s", nova.NodeID, ziv.NodeID)
	}

	msg := []byte("OpCPISync: Parallel-World Discovery from Nova -> Ziv")
	sig, err := nova.SignPayload(msg)
	if err != nil || !engine.VerifyPeerSignature(nova.NodeID, nova.SignPub, msg, sig) {
		t.Fatalf("[SEC 1] Nova Ed25519 signature verification failed: %v", err)
	}

	salt := []byte("KJV-GEN-1-1|WEB-GEN-1-1")
	kNova, err := engine.DeriveSessionKey(nova.ECDHPriv, ziv.ECDHPub, salt)
	if err != nil {
		t.Fatalf("[SEC 1] DeriveSessionKey Nova failed: %v", err)
	}
	kZiv, err := engine.DeriveSessionKey(ziv.ECDHPriv, nova.ECDHPub, salt)
	if err != nil || kNova != kZiv {
		t.Fatalf("[SEC 1] X25519 session keys diverged")
	}

	coord := engine.WitnessCoordinate{Cube: 42, Low: 100}
	genData, err := engine.SealGenesisEnvelope(msg, kNova, coord, engine.RoleKJVPrimaryEncrypt, salt)
	if err != nil {
		t.Fatalf("[SEC 1] SealGenesisEnvelope (.gen) failed: %v", err)
	}
	revData, gotCoord, gotRole, err := engine.RevealGenesisEnvelope(genData, kZiv, salt)
	if err != nil || !bytes.Equal(revData, msg) || gotCoord != coord || gotRole != engine.RoleKJVPrimaryEncrypt {
		t.Fatalf("[SEC 1] RevealGenesisEnvelope (.rev) failed: err=%v coord=%+v role=%s", err, gotCoord, gotRole)
	}
}

func TestCryptoUnit_Sec2_RejectsSpoofedNodeIDWrongKeyAndTamperedGen(t *testing.T) {
	nova, _ := engine.GenerateSovereignIdentity()
	ziv, _ := engine.GenerateSovereignIdentity()
	eve, _ := engine.GenerateSovereignIdentity()

	msg := []byte("PROTECTED-PAYLOAD")
	sig, _ := nova.SignPayload(msg)
	if engine.VerifyPeerSignature(ziv.NodeID, nova.SignPub, msg, sig) {
		t.Fatal("[SEC 2] spoofed NodeID passed VerifyPeerSignature")
	}
	if engine.VerifyPeerSignature(nova.NodeID, nova.SignPub, []byte("ALTERED"), sig) {
		t.Fatal("[SEC 2] altered message passed VerifyPeerSignature")
	}

	salt := []byte("DUAL-WITNESS")
	kNZ, _ := engine.DeriveSessionKey(nova.ECDHPriv, ziv.ECDHPub, salt)
	kEve, _ := engine.DeriveSessionKey(eve.ECDHPriv, ziv.ECDHPub, salt)

	gen, _ := engine.SealGenesisEnvelope(msg, kNZ, engine.WitnessCoordinate{Cube: 1, Low: 1}, engine.RoleKJVPrimaryEncrypt, salt)
	// Wrong session key (Eve) must fail auth!
	if _, _, _, err := engine.RevealGenesisEnvelope(gen, kEve, salt); !errors.Is(err, engine.ErrCryptoAuthFailed) {
		t.Fatalf("[SEC 2] Eve decrypted Nova->Ziv .gen! err=%v", err)
	}
	// Flipping 1 bit in header role byte (AAD!) must fail auth!
	badRole := append([]byte(nil), gen...)
	badRole[4] ^= 0x01
	if _, _, _, err := engine.RevealGenesisEnvelope(badRole, kNZ, salt); !errors.Is(err, engine.ErrCryptoAuthFailed) {
		t.Fatalf("[SEC 2] header role tamper bypassed AAD! err=%v", err)
	}
	// Invalid coordinate (>128) rejected cleanly.
	if _, err := engine.SealGenesisEnvelope(msg, kNZ, engine.WitnessCoordinate{Cube: 129, Low: 0}, engine.RoleKJVPrimaryEncrypt, salt); !errors.Is(err, engine.ErrCryptoInvalidCoord) {
		t.Fatalf("[SEC 2] expected ErrCryptoInvalidCoord, got %v", err)
	}
}

func TestCryptoUnit_Sec3_Track4VariantBytesAndSwappableKJVWEBRoles(t *testing.T) {
	nova, _ := engine.GenerateSovereignIdentity()
	ziv, _ := engine.GenerateSovereignIdentity()
	salt := []byte("TRACK4-243-PLUS-13")
	key, _ := engine.DeriveSessionKey(nova.ECDHPriv, ziv.ECDHPub, salt)

	// Verify all 13 WEB textual variant spare bytes (243..255) work cleanly!
	for v := uint16(243); v <= 255; v++ {
		coord := engine.WitnessCoordinate{Cube: 128, Low: byte(v)}
		if !coord.IsWEBVariant() {
			t.Fatalf("[SEC 3] byte %d expected IsWEBVariant=true", v)
		}
		swapped := engine.RoleKJVPrimaryEncrypt.Swap()
		if swapped != engine.RoleWEBPrimaryEncrypt || swapped.Swap() != engine.RoleKJVPrimaryEncrypt {
			t.Fatal("[SEC 3] WitnessRole.Swap() involution broken")
		}
		gen, err := engine.SealGenesisEnvelope([]byte("VARIANT-TEST"), key, coord, swapped, salt)
		if err != nil {
			t.Fatalf("[SEC 3] Seal variant %d failed: %v", v, err)
		}
		rev, gotC, gotR, err := engine.RevealGenesisEnvelope(gen, key, salt)
		if err != nil || string(rev) != "VARIANT-TEST" || gotC != coord || gotR != swapped {
			t.Fatalf("[SEC 3] Reveal variant %d failed: %v", v, err)
		}
	}
}

func TestCryptoUnit_Sec4_PreflightAndDumpCard(t *testing.T) {
	common.AssertAuditPass(t, engine.PreflightCrypto(), "PreflightCrypto")
	if !engine.IsCryptoReady() {
		t.Fatal("[SEC 4] IsCryptoReady returned false")
	}
	card := engine.ExecuteCryptoDumpCard()
	if !strings.Contains(card, ".gen <-> .rev") || !strings.Contains(card, "KJV Encrypt / WEB Decrypt") {
		t.Fatalf("[SEC 4] crypto dump card missing .gen/.rev markers:\n%s", card)
	}
}

// ==============================================================================
// BLOCK 4 & 5: AUDIT & CLOSING
// ==============================================================================
// End of tests/unit/engine/crypto_test.go
