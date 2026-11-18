// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ( Identity, Provenance, Scaling Telemetry & Architectural Boundary Contracts )
// ==============================================================================
// @project       FaithNet
// @file          engine/chunk.go
// @id            cog.engine.chunk
// @layer         Layer 1: Pure Engine Core (OSI L4 Multi-Sector Chunking)
// @archetype     pure_engine | bounded_sector_reassembler
// @version       0.2.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @updated       2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @repository    https://github.com/CreativeWorkzStudio/FaithNet
// @system        FaithNet Sovereign Adaptive Networking Substrate (FNP/1)
// @role          Bounded Multi-Sector Chunking, Out-of-Order Reassembly & Retry
// @maturity      Rev 5 Baseline | Wave 20 OSI L4 Blind-Spot Closure
// @chassis       bounded-static | Fixed 64-chunk transfer ceiling (~30 KB/transfer)
// @scale         O(N) linear split & O(1) per-chunk bitset reassembly
// @seams         in: data (FloorMaxPayload, AuditResult) | exports: ChunkReassembler
// @invariants    1. Downward Dependency Law: ZERO net, os, or external imports.
//                2. Every chunk fits strictly within 480B FloorMaxPayload (8B hdr + <=472B data).
//                3. Handles out-of-order arrivals, duplicates & selective retry lists.
//                4. Strict transfer bounds prevent memory exhaustion attacks.
//
// PURPOSE & ARCHITECTURAL ROLE:
//   Closes OSI Layer 4 Blind Spot #3: enables payloads larger than a single
//   480-byte Floppy Sector (e.g., multi-kilobyte Nova<->Ziv OpCPISync episodic
//   memories, .gen encrypted envelopes, or Cornerstone assets) to cross native
//   FNP/1 Port 7777 reliably without ever falling back to HTTP/TCP.
// ==============================================================================

package engine

// ==============================================================================
// BLOCK 2: OPENING BLOCK
// ( Chunk Framing Bounds, Error Sentinels & Bounded Reassembler Struct )
// ==============================================================================

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/CreativeWorkzStudio/FaithNet/data"
)

const (
	// ModuleIDChunk is the canonical logical ID for cog.engine.chunk.
	ModuleIDChunk = "cog.engine.chunk"

	// ChunkHeaderSize is the compact 8-byte header inside FloorMaxPayload:
	// [0..3] TransferID (uint32) | [4..5] ChunkIdx (uint16) | [6..7] TotalChunks (uint16)
	ChunkHeaderSize = 8

	// MaxChunkDataBytes is the maximum payload bytes carried per 480B sector (472B).
	MaxChunkDataBytes = data.FloorMaxPayload - ChunkHeaderSize

	// MaxChunksPerTransfer bounds a single bounded transfer (64 * 472 = 30,208 B).
	MaxChunksPerTransfer = 64

	// MaxChunkTransferBytes is the total byte ceiling per bounded transfer (30,208 B).
	MaxChunkTransferBytes = MaxChunksPerTransfer * MaxChunkDataBytes
)

var (
	// ErrChunkPayloadTooLarge signals input exceeding MaxChunkTransferBytes.
	ErrChunkPayloadTooLarge = errors.New("chunk: payload exceeds bounded transfer ceiling")
	// ErrChunkMalformed signals a truncated or invalid chunk header/index.
	ErrChunkMalformed = errors.New("chunk: malformed chunk header or out-of-range index")
	// ErrChunkTransferMismatch signals a chunk belonging to a different TransferID.
	ErrChunkTransferMismatch = errors.New("chunk: transfer ID or total count mismatch")
	// ErrChunkBufferTooSmall signals caller buffer too small for encoded chunk or reassembly.
	ErrChunkBufferTooSmall = errors.New("chunk: destination buffer too small")
)

// ChunkHeader describes one Floppy Sector fragment of a multi-sector transfer.
type ChunkHeader struct {
	TransferID  uint32
	ChunkIdx    uint16
	TotalChunks uint16
}

// ChunkReassembler reassembles up to MaxChunksPerTransfer out-of-order chunks
// into a single fixed-capacity buffer with a 64-bit arrival bitmask.
type ChunkReassembler struct {
	active      bool
	transferID  uint32
	totalChunks uint16
	receivedCnt uint16
	bitmask     uint64
	chunkLens   [MaxChunksPerTransfer]uint16
	storage     [MaxChunksPerTransfer][MaxChunkDataBytes]byte
}

// ==============================================================================
// BLOCK 3: BODY BLOCK
// ( Chunk Encode/Decode, Split Counter, Out-of-Order Reassembly & Retry List )
// ==============================================================================

// ChunkCountFor returns how many 472-byte sectors are needed to carry totalBytes.
func ChunkCountFor(totalBytes int) (uint16, error) {
	if totalBytes < 0 || totalBytes > MaxChunkTransferBytes {
		return 0, ErrChunkPayloadTooLarge
	}
	if totalBytes == 0 {
		return 1, nil
	}
	count := (totalBytes + MaxChunkDataBytes - 1) / MaxChunkDataBytes
	return uint16(count), nil
}

// EncodeChunkInto packs chunk index idx of fullPayload into dst (<= 480B) with
// zero heap allocations.
func EncodeChunkInto(dst []byte, transferID uint32, idx uint16, fullPayload []byte) (int, error) {
	total, err := ChunkCountFor(len(fullPayload))
	if err != nil {
		return 0, err
	}
	if idx >= total {
		return 0, ErrChunkMalformed
	}
	start := int(idx) * MaxChunkDataBytes
	end := start + MaxChunkDataBytes
	if end > len(fullPayload) {
		end = len(fullPayload)
	}
	dataLen := end - start
	needed := ChunkHeaderSize + dataLen
	if len(dst) < needed {
		return 0, ErrChunkBufferTooSmall
	}
	binary.BigEndian.PutUint32(dst[0:4], transferID)
	binary.BigEndian.PutUint16(dst[4:6], idx)
	binary.BigEndian.PutUint16(dst[6:8], total)
	copy(dst[ChunkHeaderSize:needed], fullPayload[start:end])
	return needed, nil
}

// DecodeChunk unpacks a raw sector payload into its ChunkHeader and data slice
// (zero-copy view into raw).
func DecodeChunk(raw []byte) (ChunkHeader, []byte, error) {
	if len(raw) < ChunkHeaderSize || len(raw) > data.FloorMaxPayload {
		return ChunkHeader{}, nil, ErrChunkMalformed
	}
	hdr := ChunkHeader{
		TransferID:  binary.BigEndian.Uint32(raw[0:4]),
		ChunkIdx:    binary.BigEndian.Uint16(raw[4:6]),
		TotalChunks: binary.BigEndian.Uint16(raw[6:8]),
	}
	if hdr.TotalChunks == 0 || hdr.TotalChunks > MaxChunksPerTransfer || hdr.ChunkIdx >= hdr.TotalChunks {
		return ChunkHeader{}, nil, ErrChunkMalformed
	}
	return hdr, raw[ChunkHeaderSize:], nil
}

// Reset clears the reassembler for a fresh transfer with zero heap allocations.
func (r *ChunkReassembler) Reset() {
	r.active = false
	r.transferID = 0
	r.totalChunks = 0
	r.receivedCnt = 0
	r.bitmask = 0
}

// Ingest absorbs one raw chunk sector (handles out-of-order and duplicate
// packets idempotently). Returns (complete, err).
func (r *ChunkReassembler) Ingest(raw []byte) (bool, error) {
	hdr, body, err := DecodeChunk(raw)
	if err != nil {
		return false, err
	}
	if !r.active {
		r.active = true
		r.transferID = hdr.TransferID
		r.totalChunks = hdr.TotalChunks
		r.receivedCnt = 0
		r.bitmask = 0
	} else if r.transferID != hdr.TransferID || r.totalChunks != hdr.TotalChunks {
		return false, ErrChunkTransferMismatch
	}

	bit := uint64(1) << hdr.ChunkIdx
	if (r.bitmask & bit) == 0 {
		copy(r.storage[hdr.ChunkIdx][:], body)
		r.chunkLens[hdr.ChunkIdx] = uint16(len(body))
		r.bitmask |= bit
		r.receivedCnt++
	}
	return r.receivedCnt == r.totalChunks, nil
}

// IsComplete reports whether every chunk of the active transfer has arrived.
func (r *ChunkReassembler) IsComplete() bool {
	return r.active && r.receivedCnt == r.totalChunks
}

// MissingIndicesInto writes all still-missing chunk indices into caller-owned
// dst for selective-ACK retry requests, returning how many chunks are missing.
func (r *ChunkReassembler) MissingIndicesInto(dst []uint16) int {
	if !r.active {
		return 0
	}
	n := 0
	for i := uint16(0); i < r.totalChunks; i++ {
		if (r.bitmask & (uint64(1) << i)) == 0 {
			if n < len(dst) {
				dst[n] = i
			}
			n++
		}
	}
	return n
}

// AssembleInto copies the complete reassembled payload sequentially into dst.
func (r *ChunkReassembler) AssembleInto(dst []byte) (int, error) {
	if !r.IsComplete() {
		return 0, ErrChunkMalformed
	}
	totalLen := 0
	for i := uint16(0); i < r.totalChunks; i++ {
		totalLen += int(r.chunkLens[i])
	}
	if len(dst) < totalLen {
		return 0, ErrChunkBufferTooSmall
	}
	off := 0
	for i := uint16(0); i < r.totalChunks; i++ {
		ln := int(r.chunkLens[i])
		copy(dst[off:off+ln], r.storage[i][:ln])
		off += ln
	}
	return totalLen, nil
}

// ==============================================================================
// BLOCK 4: AUDIT PROCEDURES BLOCK
// ==============================================================================

// PreflightChunker executes a 4-section pre-flight gate proving:
// (1) multi-sector split & reverse-order reassembly of a 2,500B payload,
// (2) rejection of malformed/oversized/mismatched chunks,
// (3) selective-retry MissingIndicesInto accuracy, and
// (4) idempotent duplicate handling & clean Reset.
//
// @triggers tests/unit/engine/chunk_test.go
func PreflightChunker() data.AuditResult {
	const check = "chunk_reassembler_preflight_4section"

	var payload [2500]byte
	for i := range payload {
		payload[i] = byte((i*31 + 7) & 0xFF)
	}
	total, err := ChunkCountFor(len(payload))
	if err != nil || total != 6 {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDChunk, check,
			fmt.Sprintf("Sec 1 ChunkCountFor(2500) got %d err=%v.", total, err))
	}

	var r ChunkReassembler
	var pkt [data.FloorMaxPayload]byte
	// Feed in reverse order (skipping chunk 2 first to verify MissingIndicesInto).
	for i := int(total) - 1; i >= 0; i-- {
		if i == 2 {
			continue
		}
		n, err := EncodeChunkInto(pkt[:], 0xFA170042, uint16(i), payload[:])
		if err != nil {
			return data.AuditFail(data.AuditSeverityFatal, ModuleIDChunk, check, err.Error())
		}
		done, err := r.Ingest(pkt[:n])
		if err != nil || done {
			return data.AuditFail(data.AuditSeverityFatal, ModuleIDChunk, check,
				"Sec 1/3 premature completion or ingest error.")
		}
	}

	// Micro Check 2: Reject mismatched TransferID & oversized payload.
	if _, err := ChunkCountFor(MaxChunkTransferBytes + 1); !errors.Is(err, ErrChunkPayloadTooLarge) {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDChunk, check,
			"Sec 2 oversized transfer bypassed ChunkCountFor.")
	}
	nBad, _ := EncodeChunkInto(pkt[:], 0xDEADBEEF, 0, payload[:])
	if _, err := r.Ingest(pkt[:nBad]); !errors.Is(err, ErrChunkTransferMismatch) {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDChunk, check,
			"Sec 2 foreign TransferID bypassed ChunkReassembler.")
	}

	// Micro Check 3: MissingIndicesInto reports exact missing chunk [2].
	var missing [8]uint16
	if m := r.MissingIndicesInto(missing[:]); m != 1 || missing[0] != 2 {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDChunk, check,
			fmt.Sprintf("Sec 3 MissingIndicesInto got %d (%v), want [2].", m, missing[:m]))
	}

	// Micro Check 4: Supply chunk 2 + duplicate chunk 2 -> exact reassembly + Reset.
	n2, _ := EncodeChunkInto(pkt[:], 0xFA170042, 2, payload[:])
	done, err := r.Ingest(pkt[:n2])
	if err != nil || !done {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDChunk, check,
			"Sec 4 final chunk did not complete transfer.")
	}
	// Duplicate must stay idempotent.
	if doneDup, errDup := r.Ingest(pkt[:n2]); errDup != nil || !doneDup {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDChunk, check,
			"Sec 4 duplicate chunk broke completion state.")
	}
	var out [2500]byte
	outN, err := r.AssembleInto(out[:])
	if err != nil || outN != len(payload) || out != payload {
		return data.AuditFail(data.AuditSeverityFatal, ModuleIDChunk, check,
			"Sec 4 reassembled payload mismatch.")
	}
	r.Reset()
	if r.IsComplete() {
		return data.AuditFail(data.AuditSeverityCritical, ModuleIDChunk, check,
			"Sec 4 Reset did not clear completion state.")
	}

	return data.AuditPass(ModuleIDChunk, check,
		"Multi-Sector Chunker passed all 4 micro gates (out-of-order, selective retry, dedup).")
}

// ==============================================================================
// BLOCK 5: CLOSING BLOCK
// ==============================================================================

// IsChunkerReady reports whether Layer 1 Multi-Sector Chunker is ready.
func IsChunkerReady() bool {
	return PreflightChunker().IsPass()
}

// ExecuteChunkerDumpCard renders a 60-column telemetry card for cog.engine.chunk.
func ExecuteChunkerDumpCard() string {
	return fmt.Sprintf(
		"+------------------------------------------------------------+\n"+
			"| FAITHNET MULTI-SECTOR TRANSPORT CHUNKER CARD (OSI L4)      |\n"+
			"+------------------------------------------------------------+\n"+
			"| Module:     %-46s |\n"+
			"| Sector Fit: 8B ChunkHdr + 472B Data = 480B Floppy MTU      |\n"+
			"| Transfer:   Up to 64 chunks (30,208 B) w/ 64-bit bitmask   |\n"+
			"| Resilience: Out-of-order, duplicate-safe & selective retry |\n"+
			"+------------------------------------------------------------+",
		ModuleIDChunk,
	)
}

// End of engine/chunk.go
