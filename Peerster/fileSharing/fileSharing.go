package fileSharing

import (
	"os"
	"io"
	"fmt"
	"sync"
	"time"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"github.com/LiangweiCHEN/Peerster/network"
	"github.com/LiangweiCHEN/Peerster/message"
	"github.com/LiangweiCHEN/Peerster/routing"
)

type FileIndexer struct {

	SharedFolder string
}

type RequestReplyChMap struct {

	Map map[string]chan *message.DataReply
	Mux sync.Mutex
}

type HashValueMap struct {

	Map map[[32]byte][]byte
	Mux sync.Mutex
}

type ChunkHashMap struct {

	Map map[string]bool
	Mux sync.Mutex
}

type FileSharer struct {

	N *network.NetworkHandler
	Indexer *FileIndexer
	RequestReplyChMap *RequestReplyChMap
	HashValueMap *HashValueMap
	HopLimit uint32
	Origin string
	RequestTimeout int
	IndexFileMap *IndexFileMap 
	ChunkHashMap *ChunkHashMap
	Dsdv *routing.DSDV
	Downloading *Downloading 
	FileLocker *FileLocker
	MetaFileMap *MetaFileMap
	ChunkMap *ChunkMap
	SearchDistributeCh chan *message.SearchRequestRelayer
	SearchReqMap *SearchReqMap
	Searcher *Searcher
}

type MetaFileMap struct {
	MetaFile map[string][]byte
	Mux sync.Mutex
}

type ChunkMap struct {
	Chunks map[string][]byte
	Mux sync.Mutex
}

type Downloading struct {
	Map map[string]chan *message.DataReply
	Mux sync.Mutex
}

type IndexFileMap struct {
	Map map[string]*IndexFile
	Mux sync.Mutex
}

type IndexFile struct {
	FileName string
	MetaHash []byte
	MetaFile []byte
	ChunkMap map[uint64]bool
	ChunkCount uint64
	Mux sync.Mutex
}

type SearchReqMap struct {
	Map map[string]bool
	Mux sync.Mutex
}
type FileLocker struct {
	Map map[string]*sync.Mutex
	Mux sync.Mutex
}

func (sharer *FileSharer) Request(hashPtr *[]byte, dest string, ch chan *message.DataReply, notification string) {
	// 1. Register requestReplyChannel and ticker
	// 2. Send request to dest
	// 3. If timeout: Resend
	// 4. If receive reply from requestReplyChannel and not empty: trigger requestChunk
	// 5. Return failure in request
	
		temp := *hashPtr
		hash := make([]byte, 0)
		for _, v := range temp {
			hash = append(hash, v)
		}

		// Step 1
		request := &message.DataRequest{

			Origin : sharer.Origin,
			Destination : dest,
			HopLimit : sharer.HopLimit,
			HashValue : hash,
		}
		gossipPacket := &message.GossipPacket{

			DataRequest : request,
		}

		replyCh := make(chan *message.DataReply)
		sharer.RequestReplyChMap.Mux.Lock()
		sharer.RequestReplyChMap.Map[dest + string(hash)] = replyCh
		sharer.RequestReplyChMap.Mux.Unlock()

		ticker := time.NewTicker(time.Duration(sharer.RequestTimeout) * time.Second)

		// Step 2
		fmt.Printf(notification)
		sharer.Dsdv.Mux.Lock()
		nextHop := sharer.Dsdv.Map[dest]
		sharer.Dsdv.Mux.Unlock()

		sharer.N.Send(gossipPacket, nextHop)

		// fmt.Printf("file request to %s\n", sharer.Dsdv.Map[dest])
		for {

			select {

			case <-ticker.C:
				// Step 3: Timeout -> resend
				fmt.Printf(notification)
				sharer.Dsdv.Mux.Lock()
				nextHop := sharer.Dsdv.Map[dest]
				sharer.Dsdv.Mux.Unlock()
				sharer.N.Send(gossipPacket, nextHop)

			case reply := <-replyCh:

				// Step 4: Break and return if empty reply
				if len(reply.Data) == 0 {

					// fmt.Println(request.HashValue)
					fmt.Printf("Peer %s does not contain value for hash %s\n", request.Destination, hex.EncodeToString(request.HashValue))
					ch<- nil
					return
				}

				// Step 5. Trigger requestChunks if reply is valid and non-empty
				hashValueArray := sha256.Sum256(reply.Data)
				if bytes.Equal(hashValueArray[:], reply.HashValue) {

					ch<- reply
					return
				} else {
					fmt.Printf("SERVER RESPOND WITH HASH VALUE %s\n", hex.EncodeToString(reply.HashValue))
					fmt.Printf("SERVER'S REPLY SUMS TO %s\n", hex.EncodeToString(hashValueArray[:]))
					fmt.Printf("SERVER RESPONSE IS INVALID\n")
					os.Exit(1)
				}
			}
		}
	
}

func (sharer *FileSharer) requestMetaFile(metahash []byte, dest string, notification string) ([]byte) {

	ch := make(chan *message.DataReply, 1)
	go sharer.Request(&metahash, dest, ch, notification)
	reply := <-ch
	close(ch)

	// Handle the destination has no corresponding metafile 
	if reply == nil {
		return nil
	}
	// Store metafile to MetaFileMap if it does not exist
	sharer.MetaFileMap.Mux.Lock()
	if _, ok := sharer.MetaFileMap.MetaFile[string(reply.HashValue)]; !ok {

		sharer.MetaFileMap.MetaFile[string(reply.HashValue)] = reply.Data
	}
	sharer.MetaFileMap.Mux.Unlock()

	// fmt.Println("reply", reply)
	if reply == nil {
		return nil
	} else {
		return reply.Data
	}
}

func (sharer *FileSharer) RequestFile(fileNamePtr *string, metahashPtr *[]byte, destPtr *string) {

		// Localize the variables 
		fileName := *fileNamePtr
		metahash := make([]byte, len(*metahashPtr))
		copy(metahash, *metahashPtr)
		dest := *destPtr

		metaNotification := fmt.Sprintf("DOWNLOADING metafile of %s from %s\n", fileName, dest)
		chunkHashes := sharer.requestMetaFile(metahash, dest, metaNotification)

		// fmt.Println("Finish metafile request")
		// Return if chunkHash is empty
		if chunkHashes == nil {
			return
		}

		// Store metahash into indexFile
		sharer.IndexFileMap.Mux.Lock()
		sharer.IndexFileMap.Map[fileName] = &IndexFile{
			FileName : fileName,
			MetaHash : metahash,
			MetaFile : chunkHashes,
			ChunkMap : make(map[uint64]bool),
			ChunkCount : uint64(len(chunkHashes) / 32),
		}
		sharer.IndexFileMap.Mux.Unlock()

		// Trigger chunk downloading handler
		sharer.Downloading.Mux.Lock()

		var downloadCh chan *message.DataReply
		if ch, ok := sharer.Downloading.Map[string(metahash)]; !ok {

			ch = make(chan *message.DataReply)
			sharer.Downloading.Map[string(metahash)] = ch
			go sharer.HandleDownloading(fileName, string(metahash), chunkHashes, ch)
			downloadCh = ch
		} else {

			downloadCh = ch
		}
		sharer.Downloading.Mux.Unlock()

		if chunkHashes != nil {

			// TODO: Modify request chunks to parallel version
			
				var wg sync.WaitGroup
				contentCh := make(chan *message.DataReply, len(chunkHashes) / 32)

				// Request chunks
				for i := 0; i < len(chunkHashes); i += 32 {

					wg.Add(1)

					// Localize chunkhash
					chunkHash := make([]byte, 0)
					for _, v := range chunkHashes[i : i + 32] {
						chunkHash = append(chunkHash, v)
					}
					
					// Request chunk with notification
					notification := fmt.Sprintf("DOWNLOADING %s chunk %d from %s\n", fileName, (i / 32) + 1,
					 dest)
					sharer.requestChunk(chunkHash, dest, contentCh, &wg, notification)
				}

				wg.Wait()

				// Put non-empty chunks into download channel
				for i := 0; i < len(chunkHashes) / 32; i += 1 {

					reply := <-contentCh
					
					if len(reply.Data) != 0 {
						downloadCh<- reply
					}
				}
			
				close(contentCh)
		}
}

func (sharer *FileSharer) requestChunk(chunkHash []byte, dest string,
										 contentCh chan *message.DataReply,
										wg *sync.WaitGroup,
										notification string) {

	ch := make(chan *message.DataReply, 1)
	defer close(ch)
	sharer.Request(&chunkHash, dest, ch, notification)
	reply := <-ch

	if reply == nil {

		wg.Done()
		fmt.Printf("Fail to request chunk with hash %s from %s\n",
					hex.EncodeToString(chunkHash),
					dest)
		os.Exit(-1)
		return
	} else {
		// Push data into channel
		contentCh<- reply

		wg.Done()
		return
	}
}


func (sharer *FileSharer) HandleReply(wrapped_pkt *message.PacketIncome) {
	// 1. Drop the reply if its chunk does not hash to hashvalue field
	// 2. Notify the requesting routine if it still exists
	// 3. Close the requestReply channel

	// Step 1
	dataReply := wrapped_pkt.Packet.DataReply
	 
	key := dataReply.Origin + string(dataReply.HashValue)

	sharer.RequestReplyChMap.Mux.Lock()
	if ch, ok := sharer.RequestReplyChMap.Map[key]; ok {

		ch<- dataReply

		// Step 2
		close(ch)
		delete(sharer.RequestReplyChMap.Map, key)
	}
	sharer.RequestReplyChMap.Mux.Unlock()

	// fmt.Printf("Receive %v from server\n", dataReply.Data)
	return
}


func (sharer *FileSharer) HandleRequest(wrapped_pkt *message.PacketIncome) {
	// 1. If find hash in metahashes, store chunks on disk and put chunk hashes
	// in chunkHashList if not done yet, return metafile
	// 2. If find chunks in chunkHashList, return chunk

	// Step 1
	dataRequest := wrapped_pkt.Packet.DataRequest
	key := string(dataRequest.HashValue)

	fmt.Printf("RECEIVE REQUEST FOR %s\n", hex.EncodeToString(dataRequest.HashValue))
	sharer.MetaFileMap.Mux.Lock()
	metafile, metaFileExist := sharer.MetaFileMap.MetaFile[key]
	sharer.MetaFileMap.Mux.Unlock()
	if metaFileExist{
		// Handle the case where a metaFile exists locally is requested
		
		// Send back metaFile
		fmt.Println("FILE BEING SEARCHED EXISTS")
		sharer.Dsdv.Mux.Lock()
		nextHop := sharer.Dsdv.Map[dataRequest.Origin]
		sharer.Dsdv.Mux.Unlock()

		sharer.N.Send(&message.GossipPacket{

			DataReply : &message.DataReply{

				Origin : sharer.Origin,
				Destination : dataRequest.Origin,
				HopLimit : sharer.HopLimit,
				HashValue : dataRequest.HashValue,
				Data : metafile,
			},
		}, nextHop)
	} else {
		// Try to obtain chunk
		sharer.ChunkMap.Mux.Lock()
		chunk, chunkExist := sharer.ChunkMap.Chunks[key]
		sharer.ChunkMap.Mux.Unlock()

		if chunkExist{

		// Send chunk back
		dataReply := &message.DataReply{

			Origin : sharer.Origin,
			Destination : dataRequest.Origin,
			HopLimit : sharer.HopLimit,
			HashValue : dataRequest.HashValue,
			Data : chunk,
		}

		sharer.Dsdv.Mux.Lock()
		nextHop := sharer.Dsdv.Map[dataReply.Destination]
		sharer.Dsdv.Mux.Unlock()
		fmt.Printf("SENDING REPLY FOR %s\n", hex.EncodeToString(dataRequest.HashValue))
		sharer.N.Send(&message.GossipPacket{

			DataReply : dataReply,
			}, nextHop)
		} else {
			// The requested stuff does not exist locally, send empty reply back

			fmt.Printf("NO DATA FOR REQUEST %s FROM %s", hex.EncodeToString(dataRequest.HashValue), dataRequest.Origin)
			dataReply := &message.DataReply{

				Origin : sharer.Origin,
				Destination : dataRequest.Origin,
				HopLimit : sharer.HopLimit,
				HashValue : dataRequest.HashValue,
				Data : make([]byte, 0),
			}

			sharer.Dsdv.Mux.Lock()
			nextHop := sharer.Dsdv.Map[dataReply.Destination]
			sharer.Dsdv.Mux.Unlock()
			sharer.N.Send(&message.GossipPacket{

				DataReply : dataReply,
			}, nextHop)
		}
	}
}

func (sharer *FileSharer) CreateIndexFile(fileNamePtr *string) (tx *message.TxPublish, err error) {

	fileName := *fileNamePtr
	tx = &message.TxPublish{
		Name : fileName,
	}
	var lock *sync.Mutex
	sharer.FileLocker.Mux.Lock()
	if _, ok := sharer.FileLocker.Map[fileName]; !ok {
		sharer.FileLocker.Map[fileName] = &sync.Mutex{}
	}
	lock = sharer.FileLocker.Map[fileName]
	sharer.FileLocker.Mux.Unlock()
	lock.Lock()

	fileName = "_SharedFiles" + "/" + fileName 
	const bufferSize = 1024 * 8

	file, err := os.Open(fileName)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	// Read chunks
	metafile := make([]byte, 0)
	totalSize := int64(0)

	for {
		buffer := make([]byte, bufferSize)
		// Read current chunk
		bytesread, err := file.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
				return tx, err
			}
			break
		}
		totalSize += int64(bytesread)

		// Compute and store hash of current chunk
		hashArray := sha256.Sum256(buffer[: bytesread])
		metafile = append(metafile, hashArray[:]...)
		// fmt.Println("The hash array has length", len(hashArray))

		// Store chunk inside sharer.ChunkMap
		hashName := string(hashArray[:])
		sharer.ChunkMap.Mux.Lock()
		sharer.ChunkMap.Chunks[hashName] = buffer[: bytesread]
		sharer.ChunkMap.Mux.Unlock()
	}

	// Compute metahash
	metaHashArray := sha256.Sum256(metafile)
	metahash := metaHashArray[:]

	// Store metafile
	metaHashName := string(metaHashArray[:])
	sharer.MetaFileMap.Mux.Lock()
	sharer.MetaFileMap.MetaFile[metaHashName] = metafile
	sharer.MetaFileMap.Mux.Unlock()

	// Build indexFile obj
	defer lock.Unlock()

	sharer.IndexFileMap.Mux.Lock()
	chunkMap := make(map[uint64]bool)
	for i := 0; i < len(metafile) / 32; i += 1 {
		chunkMap[uint64(i + 1)] = true
	}
	sharer.IndexFileMap.Map[fileName] = &IndexFile{
		FileName : fileName,
		MetaHash : metahash,
		MetaFile : metafile,
		ChunkMap : chunkMap,
		ChunkCount : uint64(len(chunkMap)),
	}
	sharer.IndexFileMap.Mux.Unlock()
	fmt.Printf("CREATE METAFILE WITH %d CHUNKS \n", len(metafile) / 32)
	fmt.Printf("CREATE INDEX FILE FOR FILE %s WITH METAHASH %s\n", fileName, hex.EncodeToString(metahash))

	// Create transaction to return
	tx.Size = totalSize
	tx.MetafileHash = metahash
    //fmt.Printf("SERVER CREATE METAFILE SUMS TO %s\n", hex.EncodeToString(metahash))
	return															  
}


func (sharer *FileSharer) HandleDownloading(fileName, metaHashStr string, chunkHashes []byte, ch chan *message.DataReply) {
	// Step 1. Construct list of hashes of chunk to download
	// Step 2. Get dataReply from channel, update storage if new
	// Step 3. Loop 2 until all chunks are downloaded
	// Step 4. Create downloaded file, stop handling downloading current file

	/* Step 1 */
	count := 0
	total := len(chunkHashes) / 32
	chunkHashStrList := make([]string, 0)
	for i := 0; i < len(chunkHashes); i += 32 {
		chunkHashStrList = append(chunkHashStrList, string(chunkHashes[i : i + 32]))
	}
	localChunkMap := make(map[string][]byte)
	sharer.IndexFileMap.Mux.Lock()
	indexFile := sharer.IndexFileMap.Map[fileName]
	metafile := indexFile.MetaFile
	sharer.IndexFileMap.Mux.Unlock()

	/* Step 2 */
	for reply := range ch {

		chunkHash := reply.HashValue
		chunkHashStr := string(chunkHash)

		sharer.ChunkMap.Mux.Lock()

		if _, ok := sharer.ChunkMap.Chunks[chunkHashStr]; !ok {

			// Store new chunk received
			count += 1
			sharer.ChunkMap.Chunks[chunkHashStr] = reply.Data
			localChunkMap[chunkHashStr] = reply.Data

			// Store new chunk's info in indexFile
			indexFile.Mux.Lock()
			for i := 0; i < len(metafile); i += 32 {
				if bytes.Compare(metafile[i : i + 32], chunkHash) == 0 {
					indexFile.ChunkMap[uint64(i + 1)] = true
				}
			}
			indexFile.Mux.Unlock()

			// Stop receiving if already received all the chunks
			if count == total {
				sharer.ChunkMap.Mux.Unlock()
				break
			}
		}
		sharer.ChunkMap.Mux.Unlock()
	}
	/* Step 4 */
	sharer.Downloading.Mux.Lock()
	delete(sharer.Downloading.Map, metaHashStr)
	close(ch)
	sharer.Downloading.Mux.Unlock()

	content := make([]byte, 0)
	for _, key := range chunkHashStrList {

		content = append(content, localChunkMap[key]...)
	}

	// Create download dir if it does not exist
	if _, err := os.Stat("_Downloads"); os.IsNotExist(err) {

		err := os.Mkdir("_Downloads", 0775)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	file, err := os.Create("_Downloads/" + fileName)
	if err != nil{
		fmt.Println(err)
		fmt.Println("invalid address")
		return
	}

	_, err = file.Write(content)
	defer file.Close()
	if err != nil {

		fmt.Println(err)
		return
	}

	fmt.Printf("RECONSTRUCTED file %s\n", fileName)
}

func (sharer *FileSharer) HandleSearchedFileDownload() {

	for wrappedRequest := range sharer.Searcher.SearchedFileDownloadCh {

		// Extract info to tell Request func
		hashPtr := &wrappedRequest.Hash
		dest := wrappedRequest.Destination
		notification := wrappedRequest.Notification
		ch := wrappedRequest.ReplyCh

		// Trigger request
		go sharer.Request(hashPtr, dest, ch, notification)

		// The reply will be directly sent to Searcher and handled by it
	}
}