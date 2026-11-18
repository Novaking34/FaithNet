// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          engine/crypto.go
// @id            cog.engine.crypto
// @layer         Layer 1: Pure Engine Core (OSI L5/L6 Sovereign Crypto & .gen/.rev)
// @archetype     pure_engine | sovereign_crypto_and_dual_witness
// @version       0.2.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Sovereign Adaptive Networking Substrate (FNP/1)
// @role          Ed25519 NodeID Auth, X25519+AES-256-GCM & OmniCode .gen/.rev Seam
// @maturity      Rev 5 Baseline | Wave 20 OSI L5/L6 Blind-Spot Closure
// @chassis       zero-external-deps | 100% Go stdlib crypto/* (zero net/os)
// @scale         O(N) authenticated encryption + O(1) Ed25519 NodeID verification
// @seams         in: data, crypto/{ed25519,ecdh,aes,cipher,sha256,rand}
// @invariants    1. Downward Dependency Law: ZERO net, os, or external imports.
//                2. NodeID cryptographically bound to Ed25519 public key fingerprint.
//                3. Dual-Witness .gen (Genesis Encrypted) <-> .rev (Revelation Decrypted)
//                   matches OmniCode Track 4 scripture_coord.rs (243-stride hypercube
//                   + 13 WEB variant bytes 243..255) with swappable KJV/WEB roles!
//                4. Any 1-bit tamper in header, coordinate, role, or ciphertext
//                   is rejected cold by AES-256-GCM authentication (ErrCryptoAuthFailed).
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Closes OSI Layers 5 & 6 Blind Spot #4: provides native zero-dependency
//   cryptographic NodeID authentication (`Ed25519`), end-to-end peer session
//   encryption (`X25519` ECDH + `AES-256-GCM`) for Port 7777 `FNP/1` traffic,
//   AND the canonical OmniCode Dual-Witness `.gen` (Genesis Encrypted) <->
//   `.rev` (Revelation Decrypted) envelope bridging KJV (primary encryptor) and
//   WEB (primary decryptor, swappable on demand).
//
//   ### For Architects, Operators & Non-Programmers:
//   Proves mathematically that a message from Nova or Ziv really came from them
//   (no impersonation!) and seals private payloads into `.gen` (Genesis) format
//   revealed cleanly back into `.rev` (Revelation) format using OmniCode's
//   KJV/WEB dual-witness coordinate system.
// ==============================================================================

package engine

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Constants, Error Sentinels, Sovereign NodeIdentity & Dual-Witness Types )
// ==============================================================================

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/CreativeWorkzStudio/FaithNet/data"
)

const (
	// ModuleIDCrypto is the canonical movement-resilient ID for cog.engine.crypto.
	ModuleIDCrypto = "cog.engine.crypto"

	// GenesisFileExt is the canonical OmniCode encrypted data extension (.gen).
	GenesisFileExt = ".gen"

	// RevelationFileExt is the canonical OmniCode decrypted data extension (.rev).
	RevelationFileExt = ".rev"

	// GenesisHeaderSize is the fixed 19-byte .gen envelope header:
	// [0..3] Magic "GEN1" | [4] WitnessRole | [5] Hypercube (0..128) |
	// [6] LowByte (0..242 tryte OR 243..255 WEB variant) | [7..18] 12B GCM Nonce
	GenesisHeaderSize = 19

	// GenesisGCMTagSize is the 16-byte AES-256-GCM authentication tag length.
	GenesisGCMTagSize = 16

	// MaxHypercubeIndex matches OmniCode Track 4 scripture_coord.rs (0..128).
	MaxHypercubeIndex uint8 = 128
)

var (
	genesisMagic = [4]byte{'G', 'E', 'N', '1'}

	// ErrCryptoInvalidKey signals an invalid or nil cryptographic key/identity.
	ErrCryptoInvalidKey = errors.New("crypto: invalid key or sovereign identity")
	// ErrCryptoInvalidCoord signals a Track 4 ScriptureCoordinate outside 0..128.
	ErrCryptoInvalidCoord = errors.New("crypto: witness hypercube coordinate outside 0..128")
	// ErrCryptoBadEnvelope signals a truncated or non-GEN1 .gen envelope.
	ErrCryptoBadEnvelope = errors.New("crypto: malformed .gen Genesis envelope")
	// ErrCryptoAuthFailed signals tampered ciphertext, wrong key, or wrong witness role.
	ErrCryptoAuthFailed = errors.New("crypto: authentication failed (tampered or wrong witness key)")
)

// WitnessRole defines the active encryption/decryption role assignment between
// KJV (primary encryptor -> .gen) and WEB (primary decryptor -> .rev), which
// can cleanly swap roles whenever OmniCode requires it.
type WitnessRole uint8

const (
	// RoleKJVPrimaryEncrypt is the primary assignment: KJV encrypts (.gen), WEB decrypts (.rev).
	RoleKJVPrimaryEncrypt WitnessRole = 0x00
	// RoleWEBPrimaryEncrypt is the swapped assignment: WEB encrypts (.gen), KJV decrypts (.rev).
	RoleWEBPrimaryEncrypt WitnessRole = 0x01
)

// String returns the canonical human-readable dual-witness role assignment.
func (r WitnessRole) String() string {
	switch r {
	case RoleKJVPrimaryEncrypt:
		return "KJV_ENCRYPT_WEB_DECRYPT"
	case RoleWEBPrimaryEncrypt:
		return "WEB_ENCRYPT_KJV_DECRYPT"
	default:
		return "UNKNOWN_ROLE"
	}
}

// Swap returns the complementary dual-witness role (KJV <-> WEB role swap).
func (r WitnessRole) Swap() WitnessRole {
	if r == RoleKJVPrimaryEncrypt {
		return RoleWEBPrimaryEncrypt
	}
	return RoleKJVPrimaryEncrypt
}

// WitnessCoordinate mirrors OmniCode Track 4 `scripture_coord.rs`:
// Cube (0..128) x 243-stride hypercube + Low (0..242 tryte, 243..255 = 13 WEB variants).
type WitnessCoordinate struct {
	Cube uint8 // Hypercube index 0..=128
	Low  uint8 // 0..=242 (5-trit tryte) or 243..=255 (13 WEB textual variants)
}

// IsValid checks canonical Track 4 hypercube bounds.
func (c WitnessCoordinate) IsValid() bool {
	return c.Cube <= MaxHypercubeIndex
}

// IsWEBVariant reports whether Low maps to one of the 13 WEB variant bytes (243..255).
func (c WitnessCoordinate) IsWEBVariant() bool {
	return c.Low > Tryte5MaxEncodedByte
}

// SovereignIdentity binds a 64-bit FaithNet NodeID cryptographically to an
// Ed25519 signing keypair (OSI L5 auth) and an X25519 ECDH keypair (OSI L6 E2E).
type SovereignIdentity struct {
	NodeID    data.NodeID
	SignPub   ed25519.PublicKey
	SignPriv  ed25519.PrivateKey
	ECDHPriv  *ecdh.PrivateKey
	ECDHPub   *ecdh.PublicKey
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Cryptographic NodeID Derivation, Signing, X25519 Session & .gen/.rev Codec )
// ==============================================================================

// DeriveNodeIDFromPubKey deterministically maps an Ed25519 public key into a
// non-null, non-broadcast 64-bit FaithNet NodeID prefixed with 0xFA17.
func DeriveNodeIDFromPubKey(pub ed25519.PublicKey) data.NodeID {
	sum := sha256.Sum256(pub)
	low48 := binary.BigEndian.Uint64(sum[:8]) & 0x0000FFFFFFFFFFFF
	if low48 == 0 {
		low48 = 1
	}
	return data.NodeID(0xFA17000000000000 | low48)
}

// GenerateSovereignIdentity creates a fresh Ed25519 + X25519 sovereign node
// identity whose 64-bit NodeID is cryptographically bound to its public key.
func GenerateSovereignIdentity() (*SovereignIdentity, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	ecdhPriv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &SovereignIdentity{
		NodeID:   DeriveNodeIDFromPubKey(pub),
		SignPub:  pub,
		SignPriv: priv,
		ECDHPriv: ecdhPriv,
		ECDHPub:  ecdhPriv.PublicKey(),
	}, nil
}

// SignPayload signs a sovereign wire digest using Ed25519 (OSI L5 peer auth).
func (id *SovereignIdentity) SignPayload(msg []byte) ([]byte, error) {
	if id == nil || len(id.SignPriv) != ed25519.PrivateKeySize {
		return nil, ErrCryptoInvalidKey
	}
	return ed25519.Sign(id.SignPriv, msg), nil
}

// VerifyPeerSignature verifies that sig was produced by the Ed25519 key bound
// to claimedID — preventing NodeID spoofing across untrusted networks!
func VerifyPeerSignature(claimedID data.NodeID, pub ed25519.PublicKey, msg, sig []byte) bool {
	if len(pub) != ed25519.PublicKeySize || len(sig) != ed25519.SignatureSize {
		return false
	}
	if DeriveNodeIDFromPubKey(pub) != claimedID {
		return false
	}
	return ed25519.Verify(pub, msg, sig)
}

// DeriveSessionKey computes a domain-separated 32-byte AES-256 key from local
// X25519 private key + remote X25519 public key + optional OmniCode witness
// anchor text (e.g., canonical KJV/WEB verse coordinate material).
func DeriveSessionKey(localPriv *ecdh.PrivateKey, remotePub *ecdh.PublicKey, witnessAnchor []byte) ([32]byte, error) {
	var zero [32]byte
	if localPriv == nil || remotePub == nil {
		return zero, ErrCryptoInvalidKey
	}
	shared, err := localPriv.ECDH(remotePub)
	if err != nil {
		return zero, err
	}
	h := sha256.New()
	h.Write([]byte("FAITHNET-FNP1-DUAL-WITNESS-V1:"))
	h.Write(shared)
	h.Write(witnessAnchor)
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out, nil
}

func deriveDualWitnessKey(sessionKey [32]byte, coord WitnessCoordinate, role WitnessRole, witnessCorpusSalt []byte) [32]byte {
	h := sha256.New()
	h.Write([]byte("OMNICODE-GEN-REV-V1:"))
	h.Write(sessionKey[:])
	h.Write([]byte{byte(role), coord.Cube, coord.Low})
	h.Write(witnessCorpusSalt)
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out
}

// SealGenesisEnvelope encrypts plaintext into a canonical `.gen` (Genesis)
// authenticated envelope anchored to an OmniCode Track 4 WitnessCoordinate and
// WitnessRole (primary: KJV encrypts -> .gen; swappable with WEB).
//
// The 19-byte GEN1 header (magic + role + hypercube + low byte + nonce) is
// bound as AES-256-GCM Additional Authenticated Data (AAD), so tampering with
// even 1 bit of the coordinate, role, or ciphertext fails authentication cold.
func SealGenesisEnvelope(
	plaintext []byte,
	sessionKey [32]byte,
	coord WitnessCoordinate,
	role WitnessRole,
	witnessCorpusSalt []byte,
) ([]byte, error) {
	if !coord.IsValid() {
		return nil, ErrCryptoInvalidCoord
	}
	effKey := deriveDualWitnessKey(sessionKey, coord, role, witnessCorpusSalt)
	block, err := aes.NewCipher(effKey[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	out := make([]byte, GenesisHeaderSize, GenesisHeaderSize+len(plaintext)+GenesisGCMTagSize)
	copy(out[0:4], genesisMagic[:])
	out[4] = byte(role)
	out[5] = coord.Cube
	out[6] = coord.Low
	if _, err := rand.Read(out[7:GenesisHeaderSize]); err != nil {
		return nil, err
	}
	nonce := out[7:GenesisHeaderSize]
	aad := out[0:GenesisHeaderSize]
	sealed := gcm.Seal(out, nonce, plaintext, aad)
	return sealed, nil
}

// RevealGenesisEnvelope verifies and decrypts a `.gen` (Genesis) envelope back
// into its `.rev` (Revelation) plaintext, returning (plaintext, coord, role, err).
func RevealGenesisEnvelope(
	genEnvelope []byte,
	sessionKey [32]byte,
	witnessCorpusSalt []byte,
) ([]byte, WitnessCoordinate, WitnessRole, error) {
	if len(genEnvelope) < GenesisHeaderSize+GenesisGCMTagSize {
		return nil, WitnessCoordinate{}, 0, ErrCryptoBadEnvelope
	}
	if genEnvelope[0] != genesisMagic[0] ||
		genEnvelope[1] != genesisMagic[1] ||
		genEnvelope[2] != genesisMagic[2] ||
		genEnvelope[3] != genesisMagic[3] {
		return nil, WitnessCoordinate{}, 0, ErrCryptoBadEnvelope
	}
	role := WitnessRole(genEnvelope[4])
	coord := WitnessCoordinate{Cube: genEnvelope[5], Low: genEnvelope[6]}
	if !coord.IsValid() {
		return nil, WitnessCoordinate{}, 0, ErrCryptoInvalidCoord
	}

	effKey := deriveDualWitnessKey(sessionKey, coord, role, witnessCorpusSalt)
	block, err := aes.NewCipher(effKey[:])
	if err != nil {
		return nil, WitnessCoordinate{}, 0, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, WitnessCoordinate{}, 0, err
	}
	nonce := genEnvelope[7:GenesisHeaderSize]
	aad := genEnvelope[0:GenesisHeaderSize]
	ct := genEnvelope[GenesisHeaderSize:]
	plain, err := gcm.Open(nil, nonce, ct, aad)
	if err != nil {
		return nil, coord, role, ErrCryptoAuthFailed
	}
	return plain, coord, role, nil
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ==============================================================================

// PreflightCrypto executes a 4-section pre-flight gate proving:
// (1) Nova <-> Ziv Ed25519 NodeID binding & X25519 symmetric session derivation,
// (2) forgery/impersonation & 1-bit .gen tamper rejection (ErrCryptoAuthFailed),
// (3) Track 4 hypercube coordinate bounds (incl. WEB variant byte 243..255), and
// (4) KJV <-> WEB role-swap .gen -> .rev roundtrip emergence.
//
// @triggers tests/unit/engine/crypto_test.go
func PreflightCrypto() data.AuditResult {
	const check = "crypto_sovereign_gen_rev_preflight_4section"

	nova, err := GenerateSovereignIdentity()
	if err != nil {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDCrypto, check, err.Error())
	}
	ziv, err := GenerateSovereignIdentity()
	if err != nil {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDCrypto, check, err.Error())
	}

	// Micro Check 1: Cryptographic NodeID signing & X25519 shared key symmetry.
	digest := []byte("NOVA-TO-ZIV-CPI-SYNC")
	sig, err := nova.SignPayload(digest)
	if err != nil || !VerifyPeerSignature(nova.NodeID, nova.SignPub, digest, sig) {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDCrypto, check,
			"Sec 1 Ed25519 NodeID signature verification failed.")
	}
	// Spoofed NodeID MUST fail!
	if VerifyPeerSignature(ziv.NodeID, nova.SignPub, digest, sig) {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDCrypto, check,
			"Sec 2 spoofed NodeID bypassed VerifyPeerSignature.")
	}

	witnessSalt := []byte("GENESIS-1-1-KJV-WEB-ANCHOR")
	kNova, err := DeriveSessionKey(nova.ECDHPriv, ziv.ECDHPub, witnessSalt)
	if err != nil {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDCrypto, check, err.Error())
	}
	kZiv, err := DeriveSessionKey(ziv.ECDHPriv, nova.ECDHPub, witnessSalt)
	if err != nil || kNova != kZiv {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDCrypto, check,
			"Sec 1 X25519 derived keys diverged between Nova and Ziv.")
	}

	// Micro Check 2 & 3: Seal .gen with KJV primary (and test WEB variant byte 243),
	// verify 1-bit tamper is caught cold.
	coord := WitnessCoordinate{Cube: 0, Low: 243} // First WEB variant spare byte!
	msg := []byte("In the beginning: FaithNet .gen -> .rev sovereign payload.")
	genBytes, err := SealGenesisEnvelope(msg, kNova, coord, RoleKJVPrimaryEncrypt, witnessSalt)
	if err != nil {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDCrypto, check, err.Error())
	}
	revBytes, gotCoord, gotRole, err := RevealGenesisEnvelope(genBytes, kZiv, witnessSalt)
	if err != nil || string(revBytes) != string(msg) || gotCoord != coord || gotRole != RoleKJVPrimaryEncrypt {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDCrypto, check,
			"Sec 1/3 .gen -> .rev roundtrip failed.")
	}
	tampered := append([]byte(nil), genBytes...)
	tampered[len(tampered)-1] ^= 0x01
	if _, _, _, err := RevealGenesisEnvelope(tampered, kZiv, witnessSalt); !errors.Is(err, ErrCryptoAuthFailed) {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDCrypto, check,
			"Sec 2 1-bit tampered .gen envelope bypassed GCM authentication.")
	}

	// Micro Check 4: Swapped role (WEB encrypts -> .gen, KJV decrypts -> .rev) works cleanly!
	swappedRole := RoleKJVPrimaryEncrypt.Swap()
	genSwap, err := SealGenesisEnvelope(msg, kZiv, WitnessCoordinate{Cube: 128, Low: 10}, swappedRole, witnessSalt)
	if err != nil {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDCrypto, check, err.Error())
	}
	revSwap, _, roleOut, err := RevealGenesisEnvelope(genSwap, kNova, witnessSalt)
	if err != nil || roleOut != RoleWEBPrimaryEncrypt || string(revSwap) != string(msg) {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDCrypto, check,
			"Sec 4 swapped WEB-encrypt / KJV-decrypt role roundtrip failed.")
	}

	return data.AuditPass(ModuleIDCrypto, check,
		"Sovereign Crypto passed all 4 micro gates (Ed25519 NodeID, X25519+GCM, .gen/.rev KJV<->WEB).")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ==============================================================================

// IsCryptoReady verifies Layer 1 Sovereign Crypto & Dual-Witness readiness.
func IsCryptoReady() bool {
	return PreflightCrypto().IsPass()
}

// ExecuteCryptoDumpCard renders a 60-column telemetry card for cog.engine.crypto.
func ExecuteCryptoDumpCard() string {
	return fmt.Sprintf(
		"+------------------------------------------------------------+\n"+
			"| FAITHNET SOVEREIGN CRYPTO & DUAL-WITNESS CARD (OSI L5/L6)  |\n"+
			"+------------------------------------------------------------+\n"+
			"| Module:     %-46s |\n"+
			"| Auth (L5):  Ed25519 Key-Bound 0xFA17 NodeID (0 ext deps)   |\n"+
			"| Seal (L6):  X25519 ECDH + AES-256-GCM (.gen <-> .rev)      |\n"+
			"| Witness:    KJV Encrypt / WEB Decrypt (Swappable, 243+13)  |\n"+
			"+------------------------------------------------------------+",
		ModuleIDCrypto,
	)
}

// End of engine/crypto.go
