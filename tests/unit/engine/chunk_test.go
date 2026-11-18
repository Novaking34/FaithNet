// ==============================================================================
// BLOCK 1: METADATA BLOCK
// ==============================================================================
// @project       FaithNet
// @file          tests/unit/engine/chunk_test.go
// @id            test.unit.engine.chunk
// @layer         Layer 4: Tier-2 Unit Verification Suite (Engine Sub-Stratum)
// @archetype     verification_suite | unit_isolation
// @version       0.2.0-dev
// @interface_rev Rev 5
// @date          2026-11-17
// @author        CreativeWorkzStudio LLC / ProfessorSeanEX
// @system        FaithNet OSI L4 Multi-Sector Chunking Unit Suite
// @role          Strict 1:1 Unit Isolation Gate for cog.engine.chunk
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

	"github.com/CreativeWorkzStudio/FaithNet/data"
	"github.com/CreativeWorkzStudio/FaithNet/engine"
	"github.com/CreativeWorkzStudio/FaithNet/tests/common"
)

// ==============================================================================
// BLOCK 3: BODY BLOCK / FOUR-SECTION ONTOLOGICAL TEST SUITE
// ==============================================================================

func TestChunkUnit_Sec1_MultiSectorOutOfOrderReassembly(t *testing.T) {
	payload := make([]byte, 5120) // 5 KB (>10 Floppy Sectors!)
	for i := range payload {
		payload[i] = byte((i*13 + 5) & 0xFF)
	}
	total, err := engine.ChunkCountFor(len(payload))
	if err != nil || total != 11 {
		t.Fatalf("[SEC 1] ChunkCountFor(5120) got %d err=%v", total, err)
	}

	var r engine.ChunkReassembler
	var sector [data.FloorMaxPayload]byte
	// Feed odd indices first, then even indices in reverse!
	for i := uint16(1); i < total; i += 2 {
		n, err := engine.EncodeChunkInto(sector[:], 0x12345678, i, payload)
		if err != nil {
			t.Fatalf("[SEC 1] EncodeChunkInto(%d) failed: %v", i, err)
		}
		if _, err := r.Ingest(sector[:n]); err != nil {
			t.Fatalf("[SEC 1] Ingest(%d) failed: %v", i, err)
		}
	}
	for i := int(total) - 1; i >= 0; i-- {
		if i%2 == 1 {
			continue
		}
		n, _ := engine.EncodeChunkInto(sector[:], 0x12345678, uint16(i), payload)
		if _, err := r.Ingest(sector[:n]); err != nil {
			t.Fatalf("[SEC 1] Ingest(%d) failed: %v", i, err)
		}
	}
	if !r.IsComplete() {
		t.Fatal("[SEC 1] reassembler not complete after all 11 chunks")
	}
	out := make([]byte, len(payload))
	gotN, err := r.AssembleInto(out)
	if err != nil || gotN != len(payload) || !bytes.Equal(out, payload) {
		t.Fatalf("[SEC 1] reassembled 5KB payload diverged (n=%d err=%v)", gotN, err)
	}
}

func TestChunkUnit_Sec2_RejectsOversizedMalformedAndForeignChunks(t *testing.T) {
	if _, err := engine.ChunkCountFor(engine.MaxChunkTransferBytes + 1); !errors.Is(err, engine.ErrChunkPayloadTooLarge) {
		t.Fatalf("[SEC 2] expected ErrChunkPayloadTooLarge, got %v", err)
	}
	if _, _, err := engine.DecodeChunk([]byte{1, 2, 3}); !errors.Is(err, engine.ErrChunkMalformed) {
		t.Fatalf("[SEC 2] expected ErrChunkMalformed for truncated chunk, got %v", err)
	}
	var r engine.ChunkReassembler
	var pkt [data.FloorMaxPayload]byte
	n1, _ := engine.EncodeChunkInto(pkt[:], 100, 0, []byte("HELLO-CHUNK"))
	_, _ = r.Ingest(pkt[:n1])
	n2, _ := engine.EncodeChunkInto(pkt[:], 999, 0, []byte("FOREIGN-ID"))
	if _, err := r.Ingest(pkt[:n2]); !errors.Is(err, engine.ErrChunkTransferMismatch) {
		t.Fatalf("[SEC 2] expected ErrChunkTransferMismatch, got %v", err)
	}
}

func TestChunkUnit_Sec3_SelectiveRetryMissingListAndZeroAllocEncode(t *testing.T) {
	payload := make([]byte, 1500) // 4 chunks
	var pkt [data.FloorMaxPayload]byte
	allocs := testing.AllocsPerRun(200, func() {
		_, _ = engine.EncodeChunkInto(pkt[:], 77, 1, payload)
	})
	if allocs != 0 {
		t.Fatalf("[SEC 3] EncodeChunkInto allocs=%.2f, want 0", allocs)
	}
	var r engine.ChunkReassembler
	n0, _ := engine.EncodeChunkInto(pkt[:], 77, 0, payload)
	_, _ = r.Ingest(pkt[:n0])
	n3, _ := engine.EncodeChunkInto(pkt[:], 77, 3, payload)
	_, _ = r.Ingest(pkt[:n3])
	var missing [8]uint16
	m := r.MissingIndicesInto(missing[:])
	if m != 2 || missing[0] != 1 || missing[1] != 2 {
		t.Fatalf("[SEC 3] MissingIndicesInto got %v (m=%d), want [1 2]", missing[:m], m)
	}
}

func TestChunkUnit_Sec4_PreflightAndDumpCard(t *testing.T) {
	common.AssertAuditPass(t, engine.PreflightChunker(), "PreflightChunker")
	if !engine.IsChunkerReady() {
		t.Fatal("[SEC 4] IsChunkerReady returned false")
	}
	if !strings.Contains(engine.ExecuteChunkerDumpCard(), "MULTI-SECTOR") {
		t.Fatal("[SEC 4] dump card missing MULTI-SECTOR marker")
	}
}

// ==============================================================================
// BLOCK 4 & 5: AUDIT & CLOSING
// ==============================================================================
// End of tests/unit/engine/chunk_test.go
