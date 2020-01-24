package fileSharing

import (
	"os"
	"sync"
	"encoding/hex"
	"bytes"
	"fmt"
	"time"
	"sort"
	"regexp"
	"strings"
	"strconv"
	"crypto/sha256"
	"github.com/LiangweiCHEN/Peerster/message"
)

type ChunkDest struct {

	ChunkIndex uint64
	Destination string
}

type FileRecord struct {

	Map map[string]FileInfo
	Mux sync.Mutex
}

type FileInfo struct {
	MetaHash []byte
	NumChunks uint64
	ChunkMap map[uint64]bool
}

type TargetMetahash struct {
	Map map[string][]byte
	Mux sync.Mutex
}

type TargetMetaFile struct {
	Map map[string][]byte
	Mux sync.Mutex
}

type Target struct {
	Map map[string][]ChunkDest
	Mux sync.Mutex
}

type Searcher struct {

	SendCh chan *message.SearchRequest
	ReplyCh chan *message.SearchReply
	Target *Target
	TargetMetahash *TargetMetahash
	TargetMetaFile *TargetMetaFile
	Threshold int
	InitBudget int
	MaxBudget int
	SearchedFileDownloadCh chan *WrappedDownloadRequest
}

type WrappedDownloadRequest struct {

	Hash []byte
	Destination string
	Notification string
	ReplyCh chan *message.DataReply
}

func (s *Searcher) Search(query []string, initBudget int) {
	// Periodically trigger search till either budget maximum reached
	// or at least threshold complete file found

	go func() {

		currentBudget := initBudget
		ticker := time.NewTicker(time.Duration(3) * time.Second)

		// Prepare search finish checking 
		finishCh := make(chan struct{})
		s.CheckSearchFinish(finishCh)
		s.ReplyCh = make(chan *message.SearchReply)
		// Trigger initial search
		request := &message.SearchRequest{
			
			Budget : uint64(currentBudget),
			Keywords : query,
		}
		s.SendCh<- request
		fmt.Printf("SEARCH WITH BUDGET %d\n", request.Budget)
		// Trigger incremental search or finish search
		for {
			select {
			case <-ticker.C:
				// Handle timeout 
				if request.Budget < uint64(s.MaxBudget) {
				
					request.Budget *= 2
					s.SendCh<- request
				} else {
					fmt.Printf("Fail to obtain enough matching files for query  %s\n", query)
				}
			case <-finishCh:
				fmt.Printf("SUCCESSFULLY FIND AT LEAST TWO FILE FOR %s\n", strings.Join(query, ","))
				s.Target.Mux.Lock()
				for k, _ := range s.TargetMetahash.Map {
					fmt.Printf("The found file are %s\n", k)
				}
				s.Target.Mux.Unlock()
				return
				// TODO: Decide what to end with search finish successfully
			}
		}
	}()
}


func (s *Searcher) CheckSearchFinish(ch chan struct{}){
	// Get result from ReplyCh
	// Update records of chunks and their location
	// Emit finish signal once at least threshold complete file found

	go func() {

		fileRecord := &FileRecord{
			Map : make(map[string]FileInfo),
		}
		finishedMap := make(map[string]bool)

		for reply := range s.ReplyCh {

			// Update fileRecord if new file or new chunk
			for _, result := range reply.Results {

				// Continue if current file has been finished searching
				fileIndex := result.FileName //+ reply.Origin
				if _, finished := finishedMap[fileIndex]; finished {
					continue
				}

				indicatingString := fmt.Sprintf("FOUND match %s at %s metafile=%s chunks=", 
										result.FileName[13: ],
										reply.Origin,
										hex.EncodeToString(result.MetafileHash))

				intChunkMap := make([]int, len(result.ChunkMap))
				strChunkMap := make([]string, len(result.ChunkMap))
				for i := range intChunkMap {
					intChunkMap[i] = int(result.ChunkMap[i])
				}
				sort.Ints(intChunkMap)
				for i := range strChunkMap {
					strChunkMap[i] = strconv.Itoa(intChunkMap[i])
				}
				indicatingString += strings.Join(strChunkMap, ",")
				fmt.Println(indicatingString)

				if fileInfo, ok := fileRecord.Map[fileIndex]; !ok {

					// Add unseen file to record
					chunkMap := make(map[uint64]bool)
					s.TargetMetahash.Mux.Lock()
					fmt.Println("I CAN OBTAIN TARGET METAHASH MUX")
					s.TargetMetahash.Map[fileIndex] = result.MetafileHash
					s.TargetMetahash.Mux.Unlock()

					s.Target.Mux.Lock()
					fmt.Println("I CAN OBTAIN TARGET MUX")
					metafilestr := string(result.MetafileHash)
					chunkDestSlice := make([]ChunkDest, 0)
					for _, index := range result.ChunkMap {
						chunkMap[index] = true

						// Add origin of chunk to Target 
						chunkDestSlice = append(chunkDestSlice, ChunkDest{
							ChunkIndex : index,
							Destination : reply.Origin,
						})
					}	
					s.Target.Map[metafilestr] = chunkDestSlice
					s.Target.Mux.Unlock()
					fileRecord.Map[fileIndex] = FileInfo{
						MetaHash : result.MetafileHash,
						NumChunks : result.ChunkCount,
						ChunkMap : chunkMap,
					}
				} else {
					// Store unseen chunks of recorded file to records
					// and store their origin in target
					chunkMap := fileInfo.ChunkMap
					s.Target.Mux.Lock()
					chunkDestSlice := s.Target.Map[string(result.MetafileHash)]
					for _, index := range result.ChunkMap {
						if _, ok := chunkMap[index]; !ok {
							chunkMap[index] = true

							chunkDestSlice = append(chunkDestSlice, ChunkDest{
								ChunkIndex : index,
								Destination : reply.Origin,
							})
						}
					}
					s.Target.Mux.Unlock()
				}

				// Check whether all the chunks of a file have been found
				count := uint64(0)
				for _ = range fileRecord.Map[fileIndex].ChunkMap {
					count += 1
				}
				if count == result.ChunkCount {
					finishedMap[fileIndex] = true
				}
			}

			// Trigger stop searching signal if at least threshold files 
			// have been completely found
			foundCount := 0
			for _ = range finishedMap {
				foundCount += 1
			}
			fmt.Printf("RECEIVE %d GUYS\n", foundCount)
			if foundCount >= s.Threshold {
				close(ch)
				s.ReplyCh = nil
				fmt.Println("SEARCH FINISHED")
				return
			}
		}
	}()
}

func (sharer *FileSharer) HandleSearch(request *message.SearchRequest, relayer string) {
	// Step 0. Disgard packet if same query have not yet timeout
	// Step 1. Distribute query to neighbours apart from relayer if budget is positive
	// Step 2. Trigger search locally
	// Step 3. Send back found files

	/* STEP 0 */
	requestKey := request.Origin + strings.Join(request.Keywords, "")
	sharer.SearchReqMap.Mux.Lock()
	if _, ok := sharer.SearchReqMap.Map[requestKey]; ok {
		fmt.Println("FREQUENT REQUEST!!!")
		return
	} else {
		sharer.SearchReqMap.Map[requestKey] = true
	}
	sharer.SearchReqMap.Mux.Unlock()

	// Start time ticker to terminate protection against frequent request attack
	go func(requestKey string) {
		ticker := time.NewTicker(time.Duration(500) * time.Millisecond)
		for {
			select {
			case <-ticker.C:
				sharer.SearchReqMap.Mux.Lock()
				delete(sharer.SearchReqMap.Map, requestKey)
				sharer.SearchReqMap.Mux.Unlock()
				return
			}
		}
	}(requestKey)

	fmt.Printf("PEERS WANT TO SEARCH FOR %s\n", strings.Join(request.Keywords, ","))
	/* STEP 1 */
	// Decrement budget by 1
	request.Budget -= 1

	// Push it to distribute channel listen by gossiper who will handle distribute issue
	if request.Budget > 0 {
		requestRelayer := &message.SearchRequestRelayer{
			SearchRequest : request,
			Relayer : relayer,
		}
		sharer.SearchDistributeCh<- requestRelayer
	}
	/* STEP 2 */
	results := make([]*message.SearchResult, 0)
	candidateFiles := make(map[string]*IndexFile)

	// Find list of files satisfying query
	sharer.IndexFileMap.Mux.Lock()
	for fileName, indexFile := range sharer.IndexFileMap.Map {
		for _, key := range request.Keywords {
			if match, _ := regexp.MatchString(".*" + key + ".*", fileName); match {
				candidateFiles[indexFile.FileName] = indexFile
				break
			}
		}
	}
	sharer.IndexFileMap.Mux.Unlock()

	/* STEP 3 */
	// Form results from matched files
	for _, indexFile := range candidateFiles {
		indexFile.Mux.Lock()
		chunkMap := make([]uint64, 0)
		for index, _ := range indexFile.ChunkMap {
			chunkMap = append(chunkMap, index)
		}
		fmt.Println(chunkMap)
		searchResult := &message.SearchResult{
			FileName : indexFile.FileName,
			MetafileHash : indexFile.MetaHash,
			ChunkMap : chunkMap,
			ChunkCount : indexFile.ChunkCount,
		}
		results = append(results, searchResult)
		indexFile.Mux.Unlock()
	}

	// Build reply
	reply := &message.SearchReply{
		Origin : sharer.Origin,
		Destination : request.Origin,
		HopLimit : sharer.HopLimit,
		Results : results,
	}

	// Find next hop 
	sharer.Dsdv.Mux.Lock()
	nextHop := sharer.Dsdv.Map[request.Origin]
	sharer.Dsdv.Mux.Unlock()

	fmt.Printf("SENDING RESPONSE WITH LEN %d TO SEARCH TO PEER\n", len(reply.Results))
	// Send reply back to requester
	sharer.N.Send(&message.GossipPacket{
		SearchReply : reply,
	}, nextHop)
}

func (s *Searcher) RequestSearchedFile(fileName string, metahash []byte) {
	// STEP 1. Download the metafile
	// STEP 2. Download the chunks according to the chunkDest

	fmt.Printf("TRYING TO DOWNLOAD %s\n", fileName)
	/* STEP 1 */
	// Get an arbitray peer containing the metafile corresponding to metahash
	metahashstr := string(metahash)
	s.Target.Mux.Lock()
	var metahashDest string
	if targetPeers, ok := s.Target.Map[metahashstr]; !ok {
		s.Target.Mux.Unlock()
		fmt.Println("CANNOT FIND METAHASH")
		return
	} else {
		metahashDest = targetPeers[0].Destination
	}
	s.Target.Mux.Unlock()
	// Send request for metafile to this peer
	replyCh := make(chan *message.DataReply)
	wrappedDownloadRequest := &WrappedDownloadRequest{
		Hash : metahash,
		Destination : metahashDest,
		Notification : fmt.Sprintf("DOWNLOADING metafile of %s from %s\n", fileName, metahashDest),
		ReplyCh : replyCh,
	}
	s.SearchedFileDownloadCh<- wrappedDownloadRequest

	// Get response from replych
	metafileReply := <-replyCh

	// Get the metafile
	metafile := metafileReply.Data

	// Check the validity of metafile
	if metahashReceived := sha256.Sum256(metafile); 
		bytes.Compare(metahashReceived[:], metahash) != 0 {

		fmt.Println("Received metahash does not match metafile")
		os.Exit(-1)
	}

	// Store metafile downloaded
	s.TargetMetaFile.Mux.Lock()
	s.TargetMetaFile.Map[metahashstr] = metafile
	s.TargetMetaFile.Mux.Unlock()

	// Trigger download of chunks
	localChunkDests := make(map[uint64]string)
	s.Target.Mux.Lock()
	for _, chunkDest := range s.Target.Map[metahashstr] {
		localChunkDests[chunkDest.ChunkIndex] = chunkDest.Destination
	}
	s.Target.Mux.Unlock()

	fmt.Println(len(metafile))
	chunkData := make(map[int][]byte)
	var chunkDataLock sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < len(metafile); i += 32 {

		// Increment wg
		wg.Add(1)

		// Localize chunkHash
		i := i
		chunkHash := metafile[i : i + 32]

		// Find peer containing the chunk
		target, ok := localChunkDests[uint64(i / 32 + 1)]
		if !ok {
			fmt.Printf("Cannot find target for %d th chunk\n",  i / 32 + 1)
			os.Exit(-1)
		}

		// Construct datarequest and replyCh
		replyCh := make(chan *message.DataReply)
		dataRequest := &WrappedDownloadRequest{

			Hash : chunkHash,
			Destination : target,
			Notification : fmt.Sprintf("DOWNLOADING %s chunk %d from %s\n",
						 fileName, i / 32 + 1, target),
			ReplyCh : replyCh,
		}

		// Trigger request
		s.SearchedFileDownloadCh<- dataRequest

		// Handle reply 
		// Disable parallel download since it leads to congestion
		//go func(replyCh chan *message.DataReply, wg *sync.WaitGroup, i int) {

		dataReply := <-replyCh

		// Put new chunk into ChunkData
		chunkDataLock.Lock()
		chunkData[i / 32] = dataReply.Data
		chunkDataLock.Unlock()

		// fmt.Printf("RECEIVE CHUNK %d WITH HASH %s\n", i / 32 + 1, hex.EncodeToString(dataReply.HashValue))
		// Increment waitgroup
		wg.Done()
		//}(replyCh, &wg, i)
	}

	// Store downloaded file onto disk if finish downloading
	wg.Wait()

	content := make([]byte, 0)
	for i := 0; i < len(metafile) / 32; i += 1 {
		content = append(content, chunkData[i]...)
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