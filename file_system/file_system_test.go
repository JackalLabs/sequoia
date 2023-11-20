package file_system

//nolint:all
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"testing"

	"github.com/JackalLabs/sequoia/proofs"
	"github.com/wealdtech/go-merkletree/v2"
	"github.com/wealdtech/go-merkletree/v2/sha3"

	"github.com/JackalLabs/sequoia/logger"
	"github.com/rs/zerolog"

	"github.com/dgraph-io/badger/v4"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
)

import _ "net/http/pprof"

var table = []struct {
	input int
}{
	{input: 1024},                    // 1kib
	{input: 1024 * 10},               // 10kib
	{input: 1024 * 1024},             // 1mib
	{input: 1024 * 1024 * 1024},      // 1gib
	{input: 1024 * 1024 * 1024 * 10}, // 10gib
}

func BenchmarkFileWrites(b *testing.B) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: ioutil.Discard})
	log.Logger = log.With().Caller().Logger()

	options := badger.DefaultOptions("/tmp/badger")
	options.Logger = &logger.SequoiaLogger{}

	db, err := badger.Open(options)
	require.NoError(b, err)

	err = db.DropAll()
	require.NoError(b, err)

	f := NewFileSystem(db)

	defer db.Close()

	go func() {
		_ = http.ListenAndServe("localhost:6060", nil)
	}()

	for _, v := range table {
		b.Run(fmt.Sprintf("input_size_%d", v.input), func(b *testing.B) {
			for i := 0; i < 1; i++ {
				token := make([]byte, v.input) // 1 kb size
				//nolint:all
				rand.Read(token)

				buf := bytes.NewBuffer(token)
				buf2 := bytes.NewBuffer(token)

				root, _, _, _, err := BuildTree(buf2, 10240)
				require.NoError(b, err)

				_, err = f.WriteFile(buf, root, "file_owner", 0, "myself", 10240)
				require.NoError(b, err)

			}
		})
	}
}

func TestWriteFile(t *testing.T) {
	db, err := badger.Open(badger.DefaultOptions("/tmp/badger"))
	require.NoError(t, err)

	err = db.DropAll()
	require.NoError(t, err)
	f := NewFileSystem(db)
	defer db.Close()

	data, err := hex.DecodeString("303030303030383597631df147918b77139b132d44798cef96879280a4b1e1309f699875c6bf57798d17bbbbe75273ba4343da20d25bbca6729ccf9b1456d0b25a08f9616a7bf414de0e15ed29f0a74378789bc7510a7d1f76348aadd93030303030383032976304f845b5c40413ec580e446491ee9bd7c780e4f2e52cb774995dcd9f10278d5ea5c5b00c2eac37039b7a844fa4a82780d9a4061a99dd1d06e130696afd07dd0e59ec275af66319a71dd53dd89f3bd6381aef3262b1bab5f8115522dbbe67411c87e827fd93d220c9d5bc60f0d55ba12df0ee3ff46ee63ecb1edf540c2aedf9c3fcf42c0310e5f7a5e69df89a0e7961e371c9f1499ccc520e283513b1e5eace184dde615078996ea67d0566b102b6f72baa9c9c76a4cc920d667f82cb55aab33c593538d636a8f1c59aa609f50eb6c20bb52c5885a7cb15cb8a3ada30a53f45ba2a3ad5c321114ffdcb8974eca8f56af3d70956af556165659b9427e078015a4fc55d6ed50a00b3aba89cd00dfdd360b5a82f631eab1be3b7c1d7eceb312733c4b21baa6640e8e5ef683a569625d8f6815858bd24a5e39f2c716862ad3cb77503e131d015f5cb615deb1974b787f85f78e85e14c92b7c8ee217a1cc997ebbb0ed3690d57a01a796692d32bb2c3c6f80af3fb104b1b506e52f94826ed6faed82df260710bb9971d1368724a7fa48c394be60d7435080dc76981c789e458a42dce0f6fe29f4e956768e0eddfff6f512a1a2e64689f82132094249df464c5286014b1835ace7b83dddea38e65e55f818ebc53d929ed38fc0997afb145c036bb1fdc7f1a2813840c69ddc1dc284d18e25b3c9b22619f0a97bcf1f36864ff0ed551e7a7249001b1f909a45b132e6de3585537240dd25941de1e4b66065626f0a2297b5c4328e6b672004e4f16aa4d742bb5b7548c4cc6756d7f2bc0de8df4fe1a21921233dd76785eb319db7bc567f2dbce5be42fdbe853edbdcf36dfbc0996874e096ea4954e4b5afb9751b0bf055778863231b4eb7a0f0839190e26db5cdd2c10f5841edc4cc85b6edf328909886d18b75e4e06210e1020fbb73b51bafdef5cd9a1bd70f52388b00a2bb555bc5e6a06bc88eeb35094a2851f3460305a83b893be857a5452b0728dae28dcd09e8e25a714cf014b557107e911fa16fa1dc6c36e4b1399cd96eca0685dc3746fa19ede15f9c0a14c5b00500b95fba05b8cb29d9c5ee6d2e164ac430e9fe56e59e10681a6f2a647c7ddf0f30ae1308035282c615c8368e")
	require.NoError(t, err)

	buf := bytes.NewBuffer(data)

	merkle := "1688dc719d1a41ff567fd54e66953f5c518044f6fed6ce814ba777b7dead4ab7d1c193448dc1c04eac05e6708dfd7a8999e9afdf6ba5c525ab7fb9c7f1e2bd4c"
	m, err := hex.DecodeString(merkle)
	require.NoError(t, err)

	size, err := f.WriteFile(buf, m, "file_owner", 0, "myself", 1024)

	require.NoError(t, err)

	require.Equal(t, len(data), size)

	err = db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			log.Info().Msg(fmt.Sprintf("key=%s", k))
		}
		return nil
	})
	require.NoError(t, err)

	ms, _, _, err := f.ListFiles()
	require.NoError(t, err)

	require.Equal(t, 1, len(ms))
}

func TestWriteAndProveFiles(t *testing.T) {
	db, err := badger.Open(badger.DefaultOptions("/tmp/badger"))
	require.NoError(t, err)
	defer db.Close()

	err = db.DropAll()
	require.NoError(t, err)
	f := NewFileSystem(db)

	size := 1024 * 1024 * 100
	var chunkSize int64 = 1024

	token := make([]byte, size)    // 1 kb size
	newToken := make([]byte, size) // 1 kb size

	//nolint:all
	rand.Read(token)
	copy(newToken, token)

	require.Equal(t, token, newToken)

	b := bytes.NewBuffer(token)
	b2 := bytes.NewBuffer(newToken)

	root, _, _, _, err := BuildTree(b, chunkSize)
	require.NoError(t, err)

	owner := "file_owner"
	var start int64 = 0

	_, err = f.WriteFile(b2, root, owner, start, "myself", chunkSize)
	require.NoError(t, err)

	err = db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			log.Info().Msg(fmt.Sprintf("key=%s", k))
		}
		return nil
	})
	require.NoError(t, err)

	ms, _, _, err := f.ListFiles()
	require.NoError(t, err)

	require.Equal(t, 1, len(ms))

	totalBlocks := size / int(chunkSize)
	for i := 0; i < totalBlocks; i++ {

		p, c, err := proofs.GenProof(f, root, owner, start, i)
		require.NoError(t, err)

		h := sha256.New()
		_, err = io.WriteString(h, fmt.Sprintf("%d%x", i, c))
		require.NoError(t, err)

		hashName := h.Sum(nil)

		var proof merkletree.Proof // unmarshal proof
		err = json.Unmarshal(p, &proof)
		require.NoError(t, err)

		verified, err := merkletree.VerifyProofUsing(hashName, false, &proof, [][]byte{root}, sha3.New512())
		require.NoError(t, err)

		require.Equal(t, true, verified)

	}

}
