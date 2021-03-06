package bigger

import (
	"strings"
	"time"
	"fmt"
	"sync"
)




type (
	coreKernel struct {
		mutex	sync.RWMutex
		blocks	map[string]*coreBlock
		indexs	[]*coreBlock
		newbies	[]coreNewbie	//当开始运行后，动态加载时的
	}
	coreNewbie struct {
		branch	string
		block	string
	}
	coreBlock struct {
		branch	string
		block	string
		chunks	[]coreChunk

		last	int
		current	int
	}
	coreChunk struct {
		time	time.Time
		name	string	//block.block
		data	Any
	}

	coreBranch	struct {
		kernel	*coreKernel
		name	string
	}
)



//注册块
func (kernel *coreKernel) chunking(branch, block string, chunk Any) (Any) {
	kernel.mutex.Lock()
	defer kernel.mutex.Unlock()

	key := fmt.Sprintf("%s.%s", branch, block)
	key = strings.ToLower(key)	//key全小写

	//先建立块
	if _,ok := kernel.blocks[key]; ok == false {
		block := &coreBlock{
			branch: branch, block: block,
			chunks: make([]coreChunk, 0),
			last: -1, current: -1,
		}
		kernel.blocks[key] = block
		kernel.indexs = append(kernel.indexs, block)
	}

	//系统正在运行的时候，记录刚刚加载的名称
	if Bigger.running {
		kernel.newbies = append(kernel.newbies, coreNewbie{
			branch: branch, block: block,
		})
	}

	info := kernel.blocks[key]
	info.chunks = append(info.chunks, coreChunk{
		time: time.Now(), name: block, data: chunk,
	})
	info.last = info.current
	info.current = len(info.chunks)-1

	return chunk
}



//返回具体的列表
func (kernel *coreKernel) chunks(branch string, prefixs ...string) ([]*coreChunk) {
	kernel.mutex.RLock()
	defer kernel.mutex.RUnlock()

	chunks := []*coreChunk{}
	//这样就有顺序了
	for _,info := range kernel.indexs {
		if info.branch == branch {
			if len(prefixs) == 0 {
				chunks = append(chunks, info.chunk())
			} else {

				for _,prefix := range prefixs {
					if strings.HasPrefix(info.block, prefix) {
						chunks = append(chunks, info.chunk())
					}
				}

			}
		}
	}

	return chunks
}



//返回版本
func (kernel *coreKernel) chunk(branch, block string) (*coreChunk) {
	kernel.mutex.RLock()
	defer kernel.mutex.RUnlock()

	key := fmt.Sprintf("%s.%s", branch, block)
	key = strings.ToLower(key)	//key全小写
	
	if info,ok := kernel.blocks[key]; ok {
		return &info.chunks[info.current]
	}

	return nil
}





//返回版本
func (block *coreBlock) chunk() (*coreChunk) {
	if len(block.chunks) > 0 {
		return &block.chunks[block.current]
	}
	return nil
}






func (branch *coreBranch) chunking(block string, chunk Any) (Any) {
	return branch.kernel.chunking(branch.name, block, chunk)
}
func (branch *coreBranch) chunks(prefixs ...string) ([]*coreChunk) {
	return branch.kernel.chunks(branch.name, prefixs...)
}
func (branch *coreBranch) chunk(block string) (*coreChunk) {
	return branch.kernel.chunk(branch.name, block)
}
func (branch *coreBranch) chunkdata(block string) (Any) {
	if chunk := branch.chunk(block); chunk != nil {
		return chunk.data
	}
	return nil
}





//上下文函数列表
func (branch *coreBranch) funcings(key string, prefixs ...string) ([]Funcing) {
	funcings := []Funcing{}
	for _,vv := range branch.chunks(prefixs...) {
		if config,ok := vv.data.(Map); ok {
			switch v:= config[key].(type) {
			case func(*Context):
				funcings = append(funcings, v)
			case []func(*Context):
				for _,vv := range v {
					funcings = append(funcings, vv)
				}
			case Funcing:
				funcings = append(funcings, v)
			case []Funcing:
				funcings = append(funcings, v...)
			default:
			}
		}
	}
    return funcings
}








//返回新加载进来的对象
func (kernel *coreKernel) lastNewbies() ([]coreNewbie) {
	kernel.mutex.RLock()
	defer kernel.mutex.RUnlock()

	newbies := kernel.newbies
	kernel.newbies = make([]coreNewbie, 0)

	return newbies
}


