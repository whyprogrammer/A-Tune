/*
 * All rights reserved by Huawei Co,.Ltd.
 */

package topology

import (
	"atune/common/cpumask"
	"atune/common/log"
	"container/list"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"

	"atune/common/utils"
)

// #cgo LDFLAGS: -lnuma
// #include <numa.h>
// import "C"

// GetNodeCpus get nodeX cpu list.
func GetNodeCpus(node int) ([]uint32, error) {
	var cpus []uint32
	path := fmt.Sprintf("/sys/devices/system/node/node%d", node)
	dir, err := os.Open(path)
	if err != nil {
		return cpus, err
	}
	defer dir.Close()
	dirs, err := dir.Readdirnames(0)
	if err != nil {
		return cpus, err
	}
	for _, dname := range dirs {
		if strings.Contains(dname, "cpu") {
			cpu, err := strconv.Atoi(dname)
			if err != nil {
				return cpus, err
			}
			cpus = append(cpus, uint32(cpu))
		}
	}
	return cpus, nil
}

const (
	sysCPUPath    = "/sys/devices/system/cpu"
	cpuOnlineFile = "online"
	coreMaskFile  = "topology/thread_siblings"
	coreIDFile    = "topology/core_id"
	numaMaskFile  = "cpumap"
	chipMaskFile  = "topology/core_siblings"
	chipIDFile    = "topology/physical_package_id"
)

const coresPerCluster = 4

// TopoType represent different cpu topo node
type TopoType uint32

const (
	// TopoTypeCPU represent single CPU thread
	TopoTypeCPU TopoType = iota
	// TopoTypeCore represent single CPU node
	TopoTypeCore
	// TopoTypeCluster represent single CPU cluster node
	TopoTypeCluster
	// TopoTypeNUMA represent single NUMA node
	TopoTypeNUMA
	// TopoTypeChip represent single physical CPU node
	TopoTypeChip
	// TopoTypeAll represent all CPUs in system
	TopoTypeAll
	// TopoTypeEnd represent TopoType total count
	TopoTypeEnd
)

// TopoLoadInfo save load info for different TopoNode
type TopoLoadInfo struct {
	bindCount int
	load      int
}

// Weight return weight of TopoLoadInfo
func (loadInfo *TopoLoadInfo) Weight() int {
	return loadInfo.load
}

// CompareLoad return compare result for src
func (loadInfo *TopoLoadInfo) CompareLoad(src *TopoLoadInfo) int {
	return loadInfo.Weight() - src.Weight()
}

// CompareBind return bind result for src
func (loadInfo *TopoLoadInfo) CompareBind(src *TopoLoadInfo) int {
	return loadInfo.bindCount - src.bindCount
}

// AddLoad load to loadInfo
func (loadInfo *TopoLoadInfo) AddLoad(load int) {
	loadInfo.load += load
}

// SubLoad load to loadInfo
func (loadInfo *TopoLoadInfo) SubLoad(load int) {
	loadInfo.load -= load
}

// GetLoad Get return load
func (loadInfo *TopoLoadInfo) GetLoad() int {
	return loadInfo.load
}

// AddBind loadInfo bindCount++
func (loadInfo *TopoLoadInfo) AddBind() {
	loadInfo.bindCount++
}

// SubBind loadInfo bindCount--
func (loadInfo *TopoLoadInfo) SubBind() {
	loadInfo.bindCount--
}

// GetBind return loadInfo bindCount
func (loadInfo *TopoLoadInfo) GetBind() int {
	return loadInfo.bindCount
}

// TopoTree represent all cpu topology in a tree
type TopoTree struct {
	typeList [TopoTypeEnd]list.List
	root     TopoNode
	numaMap  map[int]*TopoNode
}

var tree TopoTree

// TopoNode represent a node in cpu topology
type TopoNode struct {
	id         int
	topotype   TopoType
	mask       cpumask.Cpumask
	parent     *TopoNode
	loadInfo   TopoLoadInfo
	child      list.List
	typeEle    *list.Element
	AttachTask list.List
}

// Init TopoNode
func (node *TopoNode) Init() {
	node.mask.Init()
	node.parent = nil
	node.child.Init()
	node.AttachTask.Init()
}

func getMaskFromFile(mask *cpumask.Cpumask, path string) {
	line := utils.ReadAllFile(path)
	line = strings.Replace(line, ",", "", -1)
	line = strings.Replace(line, "\n", "", -1)
	_ = mask.ParseString(line)
}

func getIDFromFile(path string) int {
	line := utils.ReadAllFile(path)
	line = strings.Replace(line, "\n", "", -1)
	i, err := strconv.Atoi(line)
	if err != nil {
		fmt.Printf("get Id from %s failed\n", path)
		return -1
	}
	return i
}

func getChipID(cpu int) int {
	chipIDPath := fmt.Sprintf("%s/cpu%d/%s", sysCPUPath, cpu, chipIDFile)
	return getIDFromFile(chipIDPath)
}

func getChipMask(cpu int) *cpumask.Cpumask {
	var mask cpumask.Cpumask

	chipMaskPath := fmt.Sprintf("%s/cpu%d/%s", sysCPUPath, cpu, chipMaskFile)
	getMaskFromFile(&mask, chipMaskPath)
	return &mask
}

func getNumaID(cpu int) int {
	var node int = -1
	var err error
	cpuDirPath := fmt.Sprintf("%s/cpu%d/", sysCPUPath, cpu)

	dir, err := ioutil.ReadDir(cpuDirPath)
	if err != nil {
		return -1
	}

	for _, f := range dir {
		if strings.HasPrefix(f.Name(), "node") {
			node, err = strconv.Atoi(f.Name()[4:])
			if err != nil {
				node = -1
			}
			break
		}
	}
	return node
}

func getNumaMask(cpu int) *cpumask.Cpumask {
	var mask cpumask.Cpumask
	nodeID := getNumaID(cpu)

	if nodeID == -1 {
		return nil
	}

	numaMaskPath := fmt.Sprintf("%s/cpu%d/node%d/%s", sysCPUPath, cpu, nodeID, numaMaskFile)
	getMaskFromFile(&mask, numaMaskPath)
	return &mask
}

func getClusterID(cpu int) int {
	return cpu / coresPerCluster
}

func getClusterMask(cpu int) *cpumask.Cpumask {
	var mask cpumask.Cpumask

	mask.Init()
	cpuStart := cpu - cpu%coresPerCluster
	for i := 0; i < coresPerCluster; i++ {
		mask.Set(cpuStart + i)
	}
	return &mask
}

func getCoreID(cpu int) int {
	coreIDPath := fmt.Sprintf("%s/cpu%d/%s", sysCPUPath, cpu, coreIDFile)
	return getIDFromFile(coreIDPath)
}

func getCoreMask(cpu int) *cpumask.Cpumask {
	var mask cpumask.Cpumask

	coreMaskPath := fmt.Sprintf("%s/cpu%d/%s", sysCPUPath, cpu, coreMaskFile)
	getMaskFromFile(&mask, coreMaskPath)
	return &mask
}

func findTypeTopoNode(topotype TopoType, mask *cpumask.Cpumask) *TopoNode {
	typeHead := tree.typeList[topotype]

	for ele := typeHead.Front(); ele != nil; ele = ele.Next() {
		node, ok := (ele.Value).(*TopoNode)
		if !ok {
			return nil
		}
		if node.mask.IsEqual(mask) {
			return node
		}
	}
	return nil
}

func getChipNode(cpu int) *TopoNode {
	var node TopoNode

	chipMask := getChipMask(cpu)
	existNode := findTypeTopoNode(TopoTypeChip, chipMask)
	if existNode != nil {
		return existNode
	}

	node.Init()
	node.mask.Copy(chipMask)
	node.id = getChipID(cpu)
	node.topotype = TopoTypeChip
	return &node
}

func getNumaNode(cpu int) *TopoNode {
	var node TopoNode

	numaMask := getNumaMask(cpu)
	existNode := findTypeTopoNode(TopoTypeNUMA, numaMask)
	if existNode != nil {
		return existNode
	}

	node.Init()
	node.mask.Copy(numaMask)
	node.id = getNumaID(cpu)
	node.topotype = TopoTypeNUMA
	tree.numaMap[node.id] = &node
	return &node
}

func getClusterNode(cpu int) *TopoNode {
	var node TopoNode

	clusterMask := getClusterMask(cpu)
	existNode := findTypeTopoNode(TopoTypeCluster, clusterMask)
	if existNode != nil {
		return existNode
	}

	node.Init()
	node.mask.Copy(clusterMask)
	node.id = getClusterID(cpu)
	node.topotype = TopoTypeCluster
	return &node
}

func getCoreNode(cpu int) *TopoNode {
	var node TopoNode

	coreMask := getCoreMask(cpu)
	existNode := findTypeTopoNode(TopoTypeCore, coreMask)
	if existNode != nil {
		return existNode
	}

	node.Init()
	node.mask.Copy(coreMask)
	node.id = getCoreID(cpu)
	node.topotype = TopoTypeCore
	return &node
}

func getCPUNode(cpu int) *TopoNode {
	var node TopoNode

	node.Init()
	node.mask.Set(cpu)
	node.id = cpu
	node.topotype = TopoTypeCPU
	return &node
}

func isCPUOnline(cpu int) bool {
	onlinePath := fmt.Sprintf("%s/cpu%d/%s", sysCPUPath, cpu, cpuOnlineFile)
	line := utils.ReadAllFile(onlinePath)
	line = strings.Replace(line, "\n", "", -1)
	isOnline, err := strconv.Atoi(line)
	if err != nil {
		fmt.Println("get cpu online failed ", err)
		return false
	}
	return isOnline > 0
}

// CallBack used to tranverse TopoNode
type CallBack interface {
	Callback(node *TopoNode)
}

// ForeachChildCall tranverse TopoNode and its child
func (node *TopoNode) ForeachChildCall(fun CallBack) {
	fun.Callback(node)
	for ele := node.child.Front(); ele != nil; ele = ele.Next() {
		(ele.Value).(*TopoNode).ForeachChildCall(fun)
	}
}

type loadChangeCallback struct {
	load int
}

// Callback to update node's load
func (callback *loadChangeCallback) Callback(node *TopoNode) {
	if node.topotype == TopoTypeCPU {
		node.AddLoad(callback.load)
	}
}

// AddLoad add load to TopoNode
func (node *TopoNode) AddLoad(load int) {
	if node.topotype == TopoTypeCPU {
		p := node
		for p != nil {
			p.loadInfo.AddLoad(load)
			p = p.parent
		}
	} else {
		var callback loadChangeCallback
		callback.load = load / node.mask.Weight()
		node.ForeachChildCall(&callback)
	}
}

// SubLoad sub load from TopoNode
func (node *TopoNode) SubLoad(load int) {
	node.AddLoad(-load)
}

// SetLoad set load of TopoNode
func (node *TopoNode) SetLoad(load int) {
	node.AddLoad(load - node.GetLoad())
}

// GetLoad get load from TopoNode
func (node *TopoNode) GetLoad() int {
	return node.loadInfo.GetLoad()
}

// ID return node.id
func (node *TopoNode) ID() int {
	return node.id
}

// Type return node.topotype
func (node *TopoNode) Type() TopoType {
	return node.topotype
}

// Mask return &node.mask
func (node *TopoNode) Mask() *cpumask.Cpumask {
	return &node.mask
}

// GetNumaNodeByID return TopoNode of given numa id
func GetNumaNodeByID(id int) *TopoNode {
	return tree.numaMap[id]
}

// InitTopo initialize all TopoNodes in system
func InitTopo() {
	for _, val := range tree.typeList {
		val.Init()
	}

	tree.root.Init()
	tree.root.topotype = TopoTypeAll
	tree.typeList[TopoTypeAll].PushBack(&tree.root)
	tree.numaMap = make(map[int]*TopoNode)

	cpuNum := runtime.NumCPU()
	log.Infof("topology: cpuNum:%d\n", cpuNum)
	for cpu := 0; cpu < cpuNum; cpu++ {
		if !isCPUOnline(cpu) {
			continue
		}
		chipNode := getChipNode(cpu)
		if chipNode.parent == nil {
			chipNode.parent = &tree.root
			tree.root.child.PushBack(chipNode)
			tree.root.mask.Or(&chipNode.mask)
			chipNode.typeEle = tree.typeList[TopoTypeChip].PushBack(chipNode)
		}
		numaNode := getNumaNode(cpu)
		if numaNode.parent == nil {
			numaNode.parent = chipNode
			chipNode.child.PushBack(numaNode)
			numaNode.typeEle = tree.typeList[TopoTypeNUMA].PushBack(numaNode)
		}
		clusterNode := getClusterNode(cpu)
		if clusterNode.parent == nil {
			clusterNode.parent = numaNode
			numaNode.child.PushBack(clusterNode)
			clusterNode.typeEle = tree.typeList[TopoTypeCluster].PushBack(clusterNode)
		}
		coreNode := getCoreNode(cpu)
		if coreNode.parent == nil {
			coreNode.parent = clusterNode
			clusterNode.child.PushBack(coreNode)
			coreNode.typeEle = tree.typeList[TopoTypeCore].PushBack(coreNode)
		}
		cpuNode := getCPUNode(cpu)
		if cpuNode.parent == nil {
			cpuNode.parent = coreNode
			coreNode.child.PushBack(cpuNode)
			cpuNode.typeEle = tree.typeList[TopoTypeCPU].PushBack(cpuNode)
		}
		log.Infof("topology: cpu:%d core:%d cluster:%d numa:%d chip:%d\n", cpuNode.id, coreNode.id, clusterNode.id, numaNode.id, chipNode.id)
	}
}

// ForeachTypeCall tranverse all TopoNodes of TopoType t
func ForeachTypeCall(t TopoType, fun CallBack) {
	head := tree.typeList[t]

	for ele := head.Front(); ele != nil; ele = ele.Next() {
		fun.Callback((ele.Value).(*TopoNode))
	}
}

// ForeachType return next TopoNode of TopoType t after node
func ForeachType(t TopoType, node *TopoNode) *TopoNode {
	if node != nil {
		next := node.typeEle.Next()
		if next != nil {
			return (next.Value).(*TopoNode)
		}
		return nil
	}

	return (tree.typeList[t].Front().Value).(*TopoNode)
}

// DefaultSelectNode return top root node
func DefaultSelectNode() *TopoNode {
	return &tree.root
}

// SelectTypeIDNode return TopoNode with node.id equals id
func SelectTypeIDNode(t TopoType, id int) *TopoNode {
	for nodeEle := tree.typeList[t].Front(); nodeEle != nil; nodeEle = nodeEle.Next() {
		node, ok := (nodeEle.Value).(*TopoNode)
		if !ok {
			continue
		}
		if node.id == id {
			return node
		}
	}
	return nil
}
