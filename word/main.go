package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// 定义一个结构体来表示键值对
type kv struct {
	Key   string
	Value int
}

// 假设我们有一个英文单词列表
var words = []string{
	"Hello", "world", "this", "is", "a", "simple", "example", "of", "automatic",
	"text", "generation", "with", "Go", "language", "programming", "fun", "easy",
	// ... 添加更多单词
}

// 实现了 sort.Interface 接口的 ByValue 方法，用于对 kv 切片进行排序
type ByValue []kv

func (a ByValue) Len() int           { return len(a) }
func (a ByValue) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByValue) Less(i, j int) bool { return a[i].Value > a[j].Value } // 降序排序

// TrieNode 表示Trie树的节点
type TrieNode struct {
	children map[rune]*TrieNode
	count    int // 附加的频次字段
}

// NewTrieNode 创建一个新的Trie节点
func NewTrieNode() *TrieNode {
	return &TrieNode{
		children: make(map[rune]*TrieNode),
		count:    0,
	}
}

// Insert 插入一个单词并增加其频次（如果已存在）
func (n *TrieNode) Insert(word string, count1 int) {
	node := n
	for _, ch := range word {
		if node.children[ch] == nil {
			node.children[ch] = NewTrieNode()
		}
		node = node.children[ch]
	}
	// 到达单词的末尾，增加频次
	node.count += count1
}

// GetFrequency 获取单词的频次
func (n *TrieNode) GetFrequency(word string) int {
	node := n
	for _, ch := range word {
		if node.children[ch] == nil {
			// 单词不存在，返回0
			return 0
		}
		node = node.children[ch]
	}
	// 返回单词的频次
	return node.count
}

// init 函数用于初始化随机数生成器
func init() {
	rand.Seed(time.Now().UnixNano())
}

// generateShortText 生成一篇英文短文，大致包含 numWords 个单词
func generateShortText(numWords int) string {
	text := ""
	wordsUsed := 0
	for wordsUsed < numWords {
		// 随机选择单词
		word := words[rand.Intn(len(words))]

		// 添加标点符号或空格来模拟句子结构
		if rand.Intn(2) == 0 && wordsUsed > 0 { // 随机决定是否添加标点符号

			text += " " // 确保单词之间有空格
		} else if rand.Intn(3) == 0 && wordsUsed > 0 { // 随机决定是否开始新句子
			text += ". " // 新句子以大写字母开始，但这里为了简化我们保持小写
		} else {
			text += " " // 单词之间加空格
		}

		text += word
		wordsUsed++

		// 模拟句子长度，避免过长句子
		if rand.Intn(5) == 0 && wordsUsed > 2 { // 随机决定是否结束句子
			break
		}
	}

	// 将短文转换为首字母大写并添加句点结束
	text = strings.Title(strings.Trim(text, " .")) + "."

	return text
}
func main() {
	//从本地解析文本
	filename := "文本.text"
	text := loadText(filename)

	// 创建一个用于接收词频结果的通道
	results := make(chan map[string]int)

	// 使用sync.WaitGroup等待所有goroutines完成
	var wg sync.WaitGroup

	parts := splitString(text, 100)
	for _, part := range parts {
		wg.Add(1) //增加等待的协程数量
		go func(part string) {
			defer wg.Done()
			wordCount(part, results)
			//	进入分词方法处理part
		}(part)

	}
	go func() {
		wg.Wait()
		close(results) // 所有goroutine完成后关闭results通道
	}()
	// 合并所有词频结果
	finalResults := make(map[string]int)

	for result := range results {
		for word, count := range result {
			finalResults[word] += count
		}
		fmt.Println() // 打印空行分隔不同文本的词频
	}

	var ss []kv
	for k, v := range finalResults {
		ss = append(ss, kv{k, v})
	}
	// 对切片进行排序
	sort.Sort(ByValue(ss))
	// 创建一个Trie树的根节点
	root := NewTrieNode()
	// 打印排序后的结果
	for _, kv := range ss {
		root.Insert(kv.Key, kv.Value)
		fmt.Printf("%s: %d\n", kv.Key, kv.Value)
	}
	sum := root.GetFrequency("of")
	fmt.Println(sum)
}
func loadText(filename string) string {
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("无法读取文件 %s: %v", filename, err)
	}
	text := string(content)
	return text
}

// splitString 将字符传按照chunkSize大小进行分段

func splitString(s string, chunkSize int) []string {
	var parts []string
	runes := []rune(s) // 将字符串转换为rune切片，以支持多字节字符
	for i := 0; i < len(runes); i += chunkSize {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		parts = append(parts, s[i:end])

	}
	return parts
}

func wordCount(s string, result chan<- map[string]int) {
	//建造正则表达式  删除文本中的标点符号
	wordFreq := make(map[string]int)
	re := regexp.MustCompile(`[,.?!#$%^&*(~]+`)
	str := re.ReplaceAllString(s, "")
	words := strings.Fields(str)
	for _, word := range words {
		wordFreq[strings.ToLower(word)]++
	}
	result <- wordFreq

}
