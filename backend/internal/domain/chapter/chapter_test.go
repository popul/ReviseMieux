package chapter

import (
	"testing"
)

// Z5-AC01: canonical identity key for items
func TestZ5AC01_CanonicalKey_SameItem(t *testing.T) {
	// GIVEN two items with same term (different casing/accents), same type, same pack_id
	// WHEN we compute their canonical keys
	// THEN the keys are equal
	key1 := CanonicalKey("Vitesse", ItemKnowledge, strPtr("pack1"))
	key2 := CanonicalKey("vitesse", ItemKnowledge, strPtr("pack1"))
	if key1 != key2 {
		t.Errorf("keys should match: %q != %q", key1, key2)
	}
}

func TestZ5AC01_CanonicalKey_AccentsNormalized(t *testing.T) {
	// Accents are stripped: "énergie" == "energie"
	key1 := CanonicalKey("Énergie cinétique", ItemKnowledge, strPtr("pack1"))
	key2 := CanonicalKey("energie cinetique", ItemKnowledge, strPtr("pack1"))
	if key1 != key2 {
		t.Errorf("keys should match after accent removal: %q != %q", key1, key2)
	}
}

func TestZ5AC01_CanonicalKey_TrimWhitespace(t *testing.T) {
	key1 := CanonicalKey("  vitesse  ", ItemKnowledge, strPtr("pack1"))
	key2 := CanonicalKey("vitesse", ItemKnowledge, strPtr("pack1"))
	if key1 != key2 {
		t.Errorf("keys should match after trim: %q != %q", key1, key2)
	}
}

func TestZ5AC01_CanonicalKey_DifferentType(t *testing.T) {
	// Same term, different type → different keys
	key1 := CanonicalKey("vitesse", ItemKnowledge, strPtr("pack1"))
	key2 := CanonicalKey("vitesse", ItemProcedure, strPtr("pack1"))
	if key1 == key2 {
		t.Errorf("keys should differ for different types: %q == %q", key1, key2)
	}
}

func TestZ5AC01_CanonicalKey_DifferentPackID(t *testing.T) {
	// Same term and type, different pack_id → different keys
	key1 := CanonicalKey("vitesse", ItemKnowledge, strPtr("pack1"))
	key2 := CanonicalKey("vitesse", ItemKnowledge, strPtr("pack2"))
	if key1 == key2 {
		t.Errorf("keys should differ for different pack_id: %q == %q", key1, key2)
	}
}

func TestZ5AC01_CanonicalKey_NilPackID(t *testing.T) {
	// nil pack_id matches nil pack_id
	key1 := CanonicalKey("vitesse", ItemKnowledge, nil)
	key2 := CanonicalKey("vitesse", ItemKnowledge, nil)
	if key1 != key2 {
		t.Errorf("keys should match with nil pack_id: %q != %q", key1, key2)
	}
	// nil vs non-nil → different
	key3 := CanonicalKey("vitesse", ItemKnowledge, strPtr("pack1"))
	if key1 == key3 {
		t.Error("nil vs non-nil pack_id should differ")
	}
}

func strPtr(s string) *string { return &s }
