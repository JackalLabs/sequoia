package file_system

import "github.com/libp2p/go-libp2p/core/peer"

func (f *FileSystem) ListPeers() peer.IDSlice {
	return f.ipfsHost.Peerstore().PeersWithAddrs()
}
