package bptree

import "testing"

import (
	"bytes"
	"fmt"
)

import (
	"github.com/timtadh/fs2/consts"
)

func (t *T) assert_value(expect []byte) func(value []byte, err error) {
	return func(value []byte, err error) {
		t.assert_nil(err)
		t.assert(fmt.Sprintf("values were not equal %v != %v", expect, value), bytes.Equal(expect, value))
	}
}

func (t *T) assert_notValue(expect []byte) func(value []byte, err error) {
	return func(value []byte, err error) {
		t.assert_nil(err)
		t.assert(fmt.Sprintf("values should not have been equal %v == %v", expect, value), !bytes.Equal(expect, value))
	}
}

func TestBalanceInternal(x *testing.T) {
	t := (*T)(x)
	for TEST := 0; TEST < TESTS*10; TEST++ {
		SIZE := 1027 + TEST*16
		if SIZE > consts.BLOCKSIZE {
			SIZE = consts.BLOCKSIZE
		}
		n, err := newInternal(make([]byte, SIZE), 8)
		t.assert_nil(err)
		type KP struct {
			key []byte
			ptr uint64
		}
		make_kp := func() *KP {
			return &KP{
				key: t.rand_key(),
				ptr: t.key(t.rand_key()),
			}
		}
		kps := make([]*KP, 0, n.meta.keyCap-1)
		for i := 0; i < cap(kps); i++ {
			kp := make_kp()
			kps = append(kps, kp)
			t.assert_nil(n.putKP(kp.key, kp.ptr))
			t.assert("could not find key in internal", n.Has(kp.key))
			t.assert_ptr(kp.ptr)(n.findPtr(kp.key))
		}
		for _, kp := range kps {
			t.assert("could not find key in internal", n.Has(kp.key))
			t.assert_ptr(kp.ptr)(n.findPtr(kp.key))
		}
		b, err := newInternal(make([]byte, SIZE), 8)
		t.assert_nil(err)
		t.assert_nil(n.balance(b))
		for _, kp := range kps {
			t.assert("could not find key in internal", n.Has(kp.key) || b.Has(kp.key))
			if n.Has(kp.key) {
				t.assert_ptr(kp.ptr)(n.findPtr(kp.key))
			} else {
				t.assert_ptr(kp.ptr)(b.findPtr(kp.key))
			}
		}
		for i := 0; i < n.keyCount(); i++ {
			key := n.key(i)
			t.assert("key >= to start key in b", bytes.Compare(key, b.key(0)) < 0)
		}
	}
}

func TestBalanceLeaf(x *testing.T) {
	t := (*T)(x)
	bpt, clean := t.bpt()
	defer clean()
	for TEST := 0; TEST < TESTS; TEST++ {
		SIZE := 1027 + TEST*16
		if SIZE >= consts.BLOCKSIZE {
			SIZE = consts.BLOCKSIZE
		}
		n, err := newLeaf(0, make([]byte, SIZE), 8, 8)
		t.assert_nil(err)
		type KV struct {
			key   []byte
			value []byte
		}
		make_kv := func() *KV {
			return &KV{
				key:   t.rand_key(),
				value: t.rand_value(8),
			}
		}
		kvs := make([]*KV, 0, n.meta.keyCap/2)
		// t.Log(n)
		for i := 0; i < cap(kvs); i++ {
			kv := make_kv()
			if !n.fitsAnother() {
				break
			}
			kvs = append(kvs, kv)
			t.assert_nil(n.putKV(kv.key, kv.value))
			t.assert("could not find key in leaf", n.Has(kv.key))
			t.assert_value(kv.value)(n.firstValue(bpt.varchar, kv.key))
		}
		for _, kv := range kvs {
			t.assert("could not find key in leaf", n.Has(kv.key))
			t.assert_value(kv.value)(n.firstValue(bpt.varchar, kv.key))
		}
		b, err := newLeaf(0, make([]byte, SIZE), 8, 8)
		t.assert_nil(err)
		t.assert_nil(n.balance(b))
		for _, kv := range kvs {
			t.assert("could not find key in leaf", n.Has(kv.key) || b.Has(kv.key))
			if n.Has(kv.key) {
				t.assert_value(kv.value)(n.firstValue(bpt.varchar, kv.key))
			} else {
				t.assert_value(kv.value)(b.firstValue(bpt.varchar, kv.key))
			}
		}
		for i := 0; i < n.keyCount(); i++ {
			key := n.key(i)
			t.assert("key >= to start key in b", bytes.Compare(key, b.key(0)) < 0)
		}
	}
}

func TestBalancePureLeaf(x *testing.T) {
	t := (*T)(x)
	for TEST := 0; TEST < TESTS; TEST++ {
		SIZE := 1027 + TEST*16
		if SIZE > consts.BLOCKSIZE {
			SIZE = consts.BLOCKSIZE
		}
		n, err := newLeaf(0, make([]byte, SIZE), 8, 8)
		t.assert_nil(err)
		type KV struct {
			key   []byte
			value []byte
		}
		only_key := t.rand_key()
		make_kv := func() *KV {
			return &KV{
				key:   only_key,
				value: t.rand_value(8),
			}
		}
		kvs := make([]*KV, 0, n.meta.keyCap/2)
		// t.Log(n)
		for i := 0; i < cap(kvs); i++ {
			kv := make_kv()
			if !n.fitsAnother() {
				break
			}
			kvs = append(kvs, kv)
			t.assert_nil(n.putKV(kv.key, kv.value))
			t.assert("could not find key in leaf", n.Has(kv.key))
		}
		for _, kv := range kvs {
			t.assert("could not find key in leaf", n.Has(kv.key))
		}
		b, err := newLeaf(0, make([]byte, SIZE), 8, 8)
		t.assert_nil(err)
		t.assert_nil(n.balance(b))
		for _, kv := range kvs {
			t.assert("could not find key in leaf", n.Has(kv.key) || b.Has(kv.key))
		}
		for i := 0; i < n.keyCount(); i++ {
			key := n.key(i)
			if b.meta.keyCount > 0 {
				t.assert("key >= to start key in b", bytes.Compare(key, b.key(0)) < 0)
			}
		}
	}
}
