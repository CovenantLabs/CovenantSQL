/*
 * Copyright 2018 The CovenantSQL Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package kayak_test

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/CovenantSQL/CovenantSQL/crypto/asymmetric"
	"github.com/CovenantSQL/CovenantSQL/crypto/etls"
	"github.com/CovenantSQL/CovenantSQL/kayak"
	kt "github.com/CovenantSQL/CovenantSQL/kayak/types"
	kl "github.com/CovenantSQL/CovenantSQL/kayak/wal"
	"github.com/CovenantSQL/CovenantSQL/proto"
	crpc "github.com/CovenantSQL/CovenantSQL/rpc"
	"github.com/CovenantSQL/CovenantSQL/storage"
	"github.com/CovenantSQL/CovenantSQL/types"
	"github.com/CovenantSQL/CovenantSQL/utils"
	"github.com/CovenantSQL/CovenantSQL/utils/log"
	mock_conn "github.com/jordwest/mock-conn"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/xtaci/smux"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

type sqliteStorage struct {
	st  *storage.Storage
	dsn string
}

func newSQLiteStorage(dsn string) (s *sqliteStorage, err error) {
	s = &sqliteStorage{}
	s.st, err = storage.New(dsn)
	s.dsn = dsn
	return
}

func (s *sqliteStorage) Check(req *types.Request) (err error) {
	// no check
	return nil
}

func (s *sqliteStorage) Commit(req *types.Request, isLeader bool) (result interface{}, err error) {
	var queries []storage.Query

	for _, q := range req.Payload.Queries {
		var args []sql.NamedArg

		for _, v := range q.Args {
			args = append(args, sql.Named(v.Name, v.Value))
		}

		queries = append(queries, storage.Query{
			Pattern: q.Pattern,
			Args:    args,
		})
	}

	result, err = s.st.Exec(context.Background(), queries)

	return
}

func (s *sqliteStorage) Query(ctx context.Context, queries []storage.Query) (columns []string, types []string,
	data [][]interface{}, err error) {
	return s.st.Query(ctx, queries)
}

func (s *sqliteStorage) Close() {
	if s.st != nil {
		_ = s.st.Close()
	}
}

type fakeMux struct {
	mux map[proto.NodeID]*fakeService
}

func newFakeMux() *fakeMux {
	return &fakeMux{
		mux: make(map[proto.NodeID]*fakeService),
	}
}

func (m *fakeMux) register(nodeID proto.NodeID, s *fakeService) {
	m.mux[nodeID] = s
}

func (m *fakeMux) get(nodeID proto.NodeID) *fakeService {
	return m.mux[nodeID]
}

type fakeService struct {
	rt *kayak.Runtime
	s  *rpc.Server
}

func newFakeService(rt *kayak.Runtime) (fs *fakeService) {
	fs = &fakeService{
		rt: rt,
		s:  rpc.NewServer(),
	}

	_ = fs.s.RegisterName("Test", fs)

	return
}

func (s *fakeService) Apply(req *kt.ApplyRequest, resp *interface{}) (err error) {
	// add some delay for timeout test
	return s.rt.FollowerApply(req.Log)
}

func (s *fakeService) Fetch(req *kt.FetchRequest, resp *kt.FetchResponse) (err error) {
	var l kt.Log
	if l, err = s.rt.Fetch(req.GetContext(), req.Index); err != nil {
		return
	}

	resp.Log = l
	return
}

func (s *fakeService) serveConn(c net.Conn) {
	var r proto.NodeID
	s.s.ServeCodec(crpc.NewNodeAwareServerCodec(context.Background(), utils.GetMsgPackServerCodec(c), r.ToRawNodeID()))
}

type fakeCaller struct {
	m      *fakeMux
	target proto.NodeID
	s      *smux.Session
}

func newFakeCaller(m *fakeMux, nodeID proto.NodeID) (c *fakeCaller) {
	fakeConn := mock_conn.NewConn()
	cipher1 := etls.NewCipher([]byte("123"))
	cipher2 := etls.NewCipher([]byte("123"))
	serverConn := etls.NewConn(fakeConn.Server, cipher1, nil)
	clientConn := etls.NewConn(fakeConn.Client, cipher2, nil)

	muxSess, _ := smux.Server(serverConn, smux.DefaultConfig())
	go func() {
		for {
			s, err := muxSess.AcceptStream()
			if err != nil {
				break
			}

			go c.m.get(c.target).serveConn(s)
		}
	}()

	muxClientSess, _ := smux.Client(clientConn, smux.DefaultConfig())

	c = &fakeCaller{
		m:      m,
		target: nodeID,
		s:      muxClientSess,
	}

	return
}

func (c *fakeCaller) Call(method string, req interface{}, resp interface{}) (err error) {
	s, err := c.s.OpenStream()
	client := rpc.NewClientWithCodec(utils.GetMsgPackClientCodec(s))
	defer client.Close()

	return client.Call(method, req, resp)
}

func TestRuntime(t *testing.T) {
	Convey("runtime test", t, func(c C) {
		lvl := log.GetLevel()
		log.SetLevel(log.DebugLevel)
		defer log.SetLevel(lvl)

		db1, err := newSQLiteStorage("test1.db")
		So(err, ShouldBeNil)
		defer func() {
			db1.Close()
			os.Remove("test1.db")
		}()
		db2, err := newSQLiteStorage("test2.db")
		So(err, ShouldBeNil)
		defer func() {
			db2.Close()
			os.Remove("test2.db")
		}()

		node1 := proto.NodeID("000005aa62048f85da4ae9698ed59c14ec0d48a88a07c15a32265634e7e64ade")
		node2 := proto.NodeID("000005f4f22c06f76c43c4f48d5a7ec1309cc94030cbf9ebae814172884ac8b5")

		peers := &proto.Peers{
			PeersHeader: proto.PeersHeader{
				Leader: node1,
				Servers: []proto.NodeID{
					node1,
					node2,
				},
			},
		}

		privKey, _, err := asymmetric.GenSecp256k1KeyPair()
		So(err, ShouldBeNil)
		err = peers.Sign(privKey)
		So(err, ShouldBeNil)

		wal1 := kl.NewMemWal()
		defer wal1.Close()
		cfg1 := &kt.RuntimeConfig{
			Handler:          db1,
			PrepareThreshold: 1.0,
			CommitThreshold:  1.0,
			PrepareTimeout:   time.Second,
			CommitTimeout:    10 * time.Second,
			LogWaitTimeout:   10 * time.Second,
			Peers:            peers,
			Wal:              wal1,
			NodeID:           node1,
			ServiceName:      "Test",
			ApplyMethodName:  "Apply",
		}
		rt1, err := kayak.NewRuntime(cfg1)
		So(err, ShouldBeNil)

		wal2 := kl.NewMemWal()
		defer wal2.Close()
		cfg2 := &kt.RuntimeConfig{
			Handler:          db2,
			PrepareThreshold: 1.0,
			CommitThreshold:  1.0,
			PrepareTimeout:   time.Second,
			CommitTimeout:    10 * time.Second,
			LogWaitTimeout:   10 * time.Second,
			Peers:            peers,
			Wal:              wal2,
			NodeID:           node2,
			ServiceName:      "Test",
			ApplyMethodName:  "Apply",
		}
		rt2, err := kayak.NewRuntime(cfg2)
		So(err, ShouldBeNil)

		m := newFakeMux()
		fs1 := newFakeService(rt1)
		m.register(node1, fs1)
		fs2 := newFakeService(rt2)
		m.register(node2, fs2)

		rt1.SetCaller(node2, newFakeCaller(m, node2))
		rt2.SetCaller(node1, newFakeCaller(m, node1))

		err = rt1.Start()
		So(err, ShouldBeNil)
		defer rt1.Shutdown()

		err = rt2.Start()
		So(err, ShouldBeNil)
		defer rt2.Shutdown()

		q1 := &types.Request{
			Payload: types.RequestPayload{
				Queries: []types.Query{
					{Pattern: "CREATE TABLE IF NOT EXISTS test (t1 text, t2 text, t3 text)"},
				},
			},
		}
		So(err, ShouldBeNil)

		r1 := RandStringRunes(333)
		r2 := RandStringRunes(333)
		r3 := RandStringRunes(333)

		q2 := &types.Request{
			Payload: types.RequestPayload{
				Queries: []types.Query{
					{
						Pattern: "INSERT INTO test (t1, t2, t3) VALUES(?, ?, ?)",
						Args: []types.NamedArg{
							{Value: r1},
							{Value: r2},
							{Value: r3},
						},
					},
				},
			},
		}

		_, _, _ = rt1.Apply(context.Background(), q1)
		_, _, _ = rt2.Apply(context.Background(), q2)
		_, _, _ = rt1.Apply(context.Background(), q2)
		_, _, _, _ = db1.Query(context.Background(), []storage.Query{
			{Pattern: "SELECT * FROM test"},
		})

		var count uint64
		atomic.StoreUint64(&count, 1)

		for i := 0; i != 2000; i++ {
			atomic.AddUint64(&count, 1)
			q := &types.Request{
				Payload: types.RequestPayload{
					Queries: []types.Query{
						{
							Pattern: "INSERT INTO test (t1, t2, t3) VALUES(?, ?, ?)",
							Args: []types.NamedArg{
								{Value: r1},
								{Value: r2},
								{Value: r3},
							},
						},
					},
				},
			}

			_, _, err = rt1.Apply(context.Background(), q)
			So(err, ShouldBeNil)
		}

		// test rollback
		q := &types.Request{
			Payload: types.RequestPayload{
				Queries: []types.Query{
					{
						Pattern: "INVALID QUERY",
					},
				},
			},
		}
		_, _, err = rt1.Apply(context.Background(), q)
		So(err, ShouldNotBeNil)

		// test timeout
		q = &types.Request{
			Payload: types.RequestPayload{
				Queries: []types.Query{
					{
						Pattern: "INSERT INTO test (t1, t2, t3) VALUES(?, ?, ?)",
						Args: []types.NamedArg{
							{Value: r1},
							{Value: r2},
							{Value: r3},
						},
					},
				},
			},
		}
		cancelCtx, cancelCtxFunc := context.WithCancel(context.Background())
		cancelCtxFunc()
		_, _, err = rt1.Apply(cancelCtx, q)
		So(err, ShouldNotBeNil)

		total := atomic.LoadUint64(&count)
		_, _, d1, _ := db1.Query(context.Background(), []storage.Query{
			{Pattern: "SELECT COUNT(1) FROM test"},
		})
		So(d1, ShouldHaveLength, 1)
		So(d1[0], ShouldHaveLength, 1)
		So(fmt.Sprint(d1[0][0]), ShouldEqual, fmt.Sprint(total))

		_, _, d2, _ := db2.Query(context.Background(), []storage.Query{
			{Pattern: "SELECT COUNT(1) FROM test"},
		})
		So(d2, ShouldHaveLength, 1)
		So(d2[0], ShouldHaveLength, 1)
		So(fmt.Sprint(d2[0][0]), ShouldResemble, fmt.Sprint(total))
	})
	Convey("trivial cases", t, func() {
		node1 := proto.NodeID("000005aa62048f85da4ae9698ed59c14ec0d48a88a07c15a32265634e7e64ade")
		node2 := proto.NodeID("000005f4f22c06f76c43c4f48d5a7ec1309cc94030cbf9ebae814172884ac8b5")
		node3 := proto.NodeID("000003f49592f83d0473bddb70d543f1096b4ffed5e5f942a3117e256b7052b8")

		peers := &proto.Peers{
			PeersHeader: proto.PeersHeader{
				Leader: node1,
				Servers: []proto.NodeID{
					node1,
					node2,
				},
			},
		}

		_, err := kayak.NewRuntime(nil)
		So(err, ShouldNotBeNil)
		_, err = kayak.NewRuntime(&kt.RuntimeConfig{})
		So(err, ShouldNotBeNil)
		_, err = kayak.NewRuntime(&kt.RuntimeConfig{
			Peers: peers,
		})
		So(err, ShouldNotBeNil)

		privKey, _, err := asymmetric.GenSecp256k1KeyPair()
		So(err, ShouldBeNil)
		err = peers.Sign(privKey)
		So(err, ShouldBeNil)

		_, err = kayak.NewRuntime(&kt.RuntimeConfig{
			Peers:  peers,
			NodeID: node3,
		})
		So(err, ShouldNotBeNil)
	})
}

func TestRuntime_2(t *testing.T) {
	Convey("test log loading", t, func() {
		w, err := kl.NewLevelDBWal("testLoad.db")
		defer os.RemoveAll("testLoad.db")
		So(err, ShouldBeNil)
		err = w.Write(&kt.LogPrepare{
			LogHeader: kt.LogHeader{
				Index: 0,
				Type:  kt.LogTypePrepare,
			},
			Request: &types.Request{
				Payload: types.RequestPayload{
					Queries: []types.Query{
						{
							Pattern: "happy1",
						},
					},
				},
			},
		})
		So(err, ShouldBeNil)
		_, _ = w.Get(0)
		err = w.Write(&kt.LogPrepare{
			LogHeader: kt.LogHeader{
				Index: 1,
				Type:  kt.LogTypePrepare,
			},
			Request: &types.Request{
				Payload: types.RequestPayload{
					Queries: []types.Query{
						{
							Pattern: "happy2",
						},
					},
				},
			},
		})
		So(err, ShouldBeNil)
		err = w.Write(&kt.LogCommit{
			LogHeader: kt.LogHeader{
				Index: 2,
				Type:  kt.LogTypeCommit,
			},
			PrepareIndex:  0,
			LastCommitted: 0,
		})
		So(err, ShouldBeNil)
		err = w.Write(&kt.LogRollback{
			LogHeader: kt.LogHeader{
				Index: 3,
				Type:  kt.LogTypeRollback,
			},
			PrepareIndex: 1,
		})
		So(err, ShouldBeNil)
		w.Close()

		w, err = kl.NewLevelDBWal("testLoad.db")
		So(err, ShouldBeNil)
		defer w.Close()

		node1 := proto.NodeID("000005aa62048f85da4ae9698ed59c14ec0d48a88a07c15a32265634e7e64ade")
		peers := &proto.Peers{
			PeersHeader: proto.PeersHeader{
				Leader:  node1,
				Servers: []proto.NodeID{node1},
			},
		}

		privKey, _, err := asymmetric.GenSecp256k1KeyPair()
		So(err, ShouldBeNil)
		err = peers.Sign(privKey)
		So(err, ShouldBeNil)

		cfg := &kt.RuntimeConfig{
			Handler:          nil,
			PrepareThreshold: 1.0,
			CommitThreshold:  1.0,
			PrepareTimeout:   time.Second,
			CommitTimeout:    10 * time.Second,
			LogWaitTimeout:   10 * time.Second,
			Peers:            peers,
			Wal:              w,
			NodeID:           node1,
			ServiceName:      "Test",
			ApplyMethodName:  "Apply",
		}
		rt, err := kayak.NewRuntime(cfg)
		So(err, ShouldBeNil)

		So(rt.Start(), ShouldBeNil)
		So(func() { _ = rt.Start() }, ShouldNotPanic)

		So(rt.Shutdown(), ShouldBeNil)
		So(func() { _ = rt.Shutdown() }, ShouldNotPanic)
	})
}

func BenchmarkRuntime(b *testing.B) {
	Convey("runtime test", b, func(c C) {
		log.SetLevel(log.FatalLevel)
		f, err := os.OpenFile("test.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		So(err, ShouldBeNil)
		log.SetOutput(f)
		defer f.Close()

		db1, err := newSQLiteStorage("test1.db")
		So(err, ShouldBeNil)
		defer func() {
			db1.Close()
			_ = os.Remove("test1.db")
		}()
		db2, err := newSQLiteStorage("test2.db")
		So(err, ShouldBeNil)
		defer func() {
			db2.Close()
			_ = os.Remove("test2.db")
		}()

		node1 := proto.NodeID("000005aa62048f85da4ae9698ed59c14ec0d48a88a07c15a32265634e7e64ade")
		node2 := proto.NodeID("000005f4f22c06f76c43c4f48d5a7ec1309cc94030cbf9ebae814172884ac8b5")

		peers := &proto.Peers{
			PeersHeader: proto.PeersHeader{
				Leader: node1,
				Servers: []proto.NodeID{
					node1,
					node2,
				},
			},
		}

		privKey, _, err := asymmetric.GenSecp256k1KeyPair()
		So(err, ShouldBeNil)
		err = peers.Sign(privKey)
		So(err, ShouldBeNil)

		wal1 := kl.NewMemWal()
		defer wal1.Close()
		cfg1 := &kt.RuntimeConfig{
			Handler:          db1,
			PrepareThreshold: 1.0,
			CommitThreshold:  0.0,
			PrepareTimeout:   time.Second,
			CommitTimeout:    10 * time.Second,
			LogWaitTimeout:   10 * time.Second,
			Peers:            peers,
			Wal:              wal1,
			NodeID:           node1,
			ServiceName:      "Test",
			ApplyMethodName:  "Apply",
		}
		rt1, err := kayak.NewRuntime(cfg1)
		So(err, ShouldBeNil)

		wal2 := kl.NewMemWal()
		defer wal2.Close()
		cfg2 := &kt.RuntimeConfig{
			Handler:          db2,
			PrepareThreshold: 1.0,
			CommitThreshold:  0.0,
			PrepareTimeout:   time.Second,
			CommitTimeout:    10 * time.Second,
			LogWaitTimeout:   10 * time.Second,
			Peers:            peers,
			Wal:              wal2,
			NodeID:           node2,
			ServiceName:      "Test",
			ApplyMethodName:  "Apply",
		}
		rt2, err := kayak.NewRuntime(cfg2)
		So(err, ShouldBeNil)

		m := newFakeMux()
		fs1 := newFakeService(rt1)
		m.register(node1, fs1)
		fs2 := newFakeService(rt2)
		m.register(node2, fs2)

		rt1.SetCaller(node2, newFakeCaller(m, node2))
		rt2.SetCaller(node1, newFakeCaller(m, node1))

		err = rt1.Start()
		So(err, ShouldBeNil)
		defer rt1.Shutdown()

		err = rt2.Start()
		So(err, ShouldBeNil)
		defer rt2.Shutdown()

		q1 := &types.Request{
			Payload: types.RequestPayload{
				Queries: []types.Query{
					{Pattern: "CREATE TABLE IF NOT EXISTS test (t1 text, t2 text, t3 text)"},
				},
			},
		}
		So(err, ShouldBeNil)

		r1 := RandStringRunes(333)
		r2 := RandStringRunes(333)
		r3 := RandStringRunes(333)

		q2 := &types.Request{
			Payload: types.RequestPayload{
				Queries: []types.Query{
					{
						Pattern: "INSERT INTO test (t1, t2, t3) VALUES(?, ?, ?)",
						Args: []types.NamedArg{
							{Value: r1},
							{Value: r2},
							{Value: r3},
						},
					},
				},
			},
		}

		rt1.Apply(context.Background(), q1)
		rt2.Apply(context.Background(), q2)
		rt1.Apply(context.Background(), q2)
		db1.Query(context.Background(), []storage.Query{
			{Pattern: "SELECT * FROM test"},
		})

		b.ResetTimer()

		var count uint64
		atomic.StoreUint64(&count, 1)

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				atomic.AddUint64(&count, 1)
				q := &types.Request{
					Payload: types.RequestPayload{
						Queries: []types.Query{
							{
								Pattern: "INSERT INTO test (t1, t2, t3) VALUES(?, ?, ?)",
								Args: []types.NamedArg{
									{Value: r1},
									{Value: r2},
									{Value: r3},
								},
							},
						},
					},
				}
				_ = err
				//c.So(err, ShouldBeNil)

				_, _, err = rt1.Apply(context.Background(), q)
				//c.So(err, ShouldBeNil)
			}
		})

		b.StopTimer()

		total := atomic.LoadUint64(&count)
		_, _, d1, _ := db1.Query(context.Background(), []storage.Query{
			{Pattern: "SELECT COUNT(1) FROM test"},
		})
		So(d1, ShouldHaveLength, 1)
		So(d1[0], ShouldHaveLength, 1)
		_ = total
		//So(fmt.Sprint(d1[0][0]), ShouldEqual, fmt.Sprint(total))

		//_, _, d2, _ := db2.Query(context.Background(), []storage.Query{
		//	{Pattern: "SELECT COUNT(1) FROM test"},
		//})
		//So(d2, ShouldHaveLength, 1)
		//So(d2[0], ShouldHaveLength, 1)
		//So(fmt.Sprint(d2[0][0]), ShouldResemble, fmt.Sprint(total))

		b.StartTimer()
	})
}
