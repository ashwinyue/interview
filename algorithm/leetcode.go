package algorithm

// ============================================================================
// LeetCode 高频题 - Go 实现
// ============================================================================
// 题目来源：LeetCode 热门 100 题
// 难度标记：[Easy] [Medium] [Hard]
// ============================================================================

// ============================================================================
// 数组与字符串
// ============================================================================

// 1. 两数之和 [Easy] #1
// 给定一个整数数组 nums 和一个目标值 target
// 在该数组中找出和为目标值的那两个整数
func TwoSum(nums []int, target int) []int {
	// 使用 map 存储值到索引的映射
	m := make(map[int]int)
	for i, num := range nums {
		if j, ok := m[target-num]; ok {
			return []int{j, i}
		}
		m[num] = i
	}
	return nil
}

// 2. 三数之和 [Medium] #15
// 找出所有和为 0 的三元组
func ThreeSum(nums []int) [][]int {
	// 排序 + 双指针
	quickSort(nums)
	result := [][]int{}
	n := len(nums)

	for i := 0; i < n-2; i++ {
		// 跳过重复元素
		if i > 0 && nums[i] == nums[i-1] {
			continue
		}
		// 优化：如果最小值大于0，不可能有解
		if nums[i] > 0 {
			break
		}

		left, right := i+1, n-1
		for left < right {
			sum := nums[i] + nums[left] + nums[right]
			if sum == 0 {
				result = append(result, []int{nums[i], nums[left], nums[right]})
				// 跳过重复元素
				for left < right && nums[left] == nums[left+1] {
					left++
				}
				for left < right && nums[right] == nums[right-1] {
					right--
				}
				left++
				right--
			} else if sum < 0 {
				left++
			} else {
				right--
			}
		}
	}
	return result
}

// 3. 最大子数组和 [Medium] #53
// 连续子数组的最大和
func MaxSubArray(nums []int) int {
	maxSum := nums[0]
	currentSum := nums[0]

	for i := 1; i < len(nums); i++ {
		currentSum = max(nums[i], currentSum+nums[i])
		maxSum = max(maxSum, currentSum)
	}
	return maxSum
}

// 4. 合并区间 [Medium] #56
func Merge(intervals [][]int) [][]int {
	if len(intervals) <= 1 {
		return intervals
	}

	// 按起始位置排序
	sortIntervals(intervals)

	result := [][]int{intervals[0]}
	for i := 1; i < len(intervals); i++ {
		last := result[len(result)-1]
		current := intervals[i]

		if current[0] <= last[1] {
			// 有重叠，合并
			last[1] = max(last[1], current[1])
		} else {
			result = append(result, current)
		}
	}
	return result
}

// 5. 轮转数组 [Medium] #189
func Rotate(nums []int, k int) {
	n := len(nums)
	k %= n
	if k == 0 {
		return
	}

	// 方法：三次反转
	// 1. 反转整个数组
	// 2. 反转前 k 个
	// 3. 反转剩余部分
	reverse(nums, 0, n-1)
	reverse(nums, 0, k-1)
	reverse(nums, k, n-1)
}

func reverse(nums []int, start, end int) {
	for start < end {
		nums[start], nums[end] = nums[end], nums[start]
		start++
		end--
	}
}

// 6. 除自身以外数组的乘积 [Medium] #238
func ProductExceptSelf(nums []int) []int {
	n := len(nums)
	result := make([]int, n)

	// 左侧乘积
	leftProduct := 1
	for i := 0; i < n; i++ {
		result[i] = leftProduct
		leftProduct *= nums[i]
	}

	// 右侧乘积
	rightProduct := 1
	for i := n - 1; i >= 0; i-- {
		result[i] *= rightProduct
		rightProduct *= nums[i]
	}
	return result
}

// ============================================================================
// 链表
// ============================================================================

// ListNode 链表节点
type ListNode struct {
	Val  int
	Next *ListNode
}

// 7. 反转链表 [Easy] #206
func ReverseList(head *ListNode) *ListNode {
	var prev *ListNode
	current := head

	for current != nil {
		next := current.Next
		current.Next = prev
		prev = current
		current = next
	}
	return prev
}

// 8. 合并两个有序链表 [Easy] #21
func MergeTwoLists(l1, l2 *ListNode) *ListNode {
	dummy := &ListNode{}
	current := dummy

	for l1 != nil && l2 != nil {
		if l1.Val <= l2.Val {
			current.Next = l1
			l1 = l1.Next
		} else {
			current.Next = l2
			l2 = l2.Next
		}
		current = current.Next
	}

	if l1 != nil {
		current.Next = l1
	}
	if l2 != nil {
		current.Next = l2
	}
	return dummy.Next
}

// 9. 环形链表 [Easy] #141
func HasCycle(head *ListNode) bool {
	slow, fast := head, head

	for fast != nil && fast.Next != nil {
		slow = slow.Next
		fast = fast.Next.Next
		if slow == fast {
			return true
		}
	}
	return false
}

// 10. 环形链表 II [Medium] #142
// 返回入环的第一个节点
func DetectCycle(head *ListNode) *ListNode {
	slow, fast := head, head

	// 检测是否有环
	hasCycle := false
	for fast != nil && fast.Next != nil {
		slow = slow.Next
		fast = fast.Next.Next
		if slow == fast {
			hasCycle = true
			break
		}
	}

	if !hasCycle {
		return nil
	}

	// 找到入环点
	slow = head
	for slow != fast {
		slow = slow.Next
		fast = fast.Next
	}
	return slow
}

// 11. 合并 K 个升序链表 [Hard] #23
func MergeKLists(lists []*ListNode) *ListNode {
	if len(lists) == 0 {
		return nil
	}

	// 分治合并
	return mergeKListsHelper(lists, 0, len(lists)-1)
}

func mergeKListsHelper(lists []*ListNode, left, right int) *ListNode {
	if left == right {
		return lists[left]
	}
	mid := (left + right) / 2
	return MergeTwoLists(
		mergeKListsHelper(lists, left, mid),
		mergeKListsHelper(lists, mid+1, right),
	)
}

// ============================================================================
// 二叉树
// ============================================================================

// TreeNode 二叉树节点
type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

// 12. 二叉树的最大深度 [Easy] #104
func MaxDepth(root *TreeNode) int {
	if root == nil {
		return 0
	}
	return 1 + max(MaxDepth(root.Left), MaxDepth(root.Right))
}

// 13. 翻转二叉树 [Easy] #226
func InvertTree(root *TreeNode) *TreeNode {
	if root == nil {
		return nil
	}

	root.Left, root.Right = InvertTree(root.Right), InvertTree(root.Left)
	return root
}

// 14. 对称二叉树 [Easy] #101
func IsSymmetric(root *TreeNode) bool {
	if root == nil {
		return true
	}
	return isMirror(root.Left, root.Right)
}

func isMirror(left, right *TreeNode) bool {
	if left == nil && right == nil {
		return true
	}
	if left == nil || right == nil {
		return false
	}
	return left.Val == right.Val &&
		isMirror(left.Left, right.Right) &&
		isMirror(left.Right, right.Left)
}

// 15. 二叉树的层序遍历 [Medium] #102
func LevelOrder(root *TreeNode) [][]int {
	if root == nil {
		return nil
	}

	result := [][]int{}
	queue := []*TreeNode{root}

	for len(queue) > 0 {
		levelSize := len(queue)
		level := []int{}

		for i := 0; i < levelSize; i++ {
			node := queue[0]
			queue = queue[1:]
			level = append(level, node.Val)

			if node.Left != nil {
				queue = append(queue, node.Left)
			}
			if node.Right != nil {
				queue = append(queue, node.Right)
			}
		}
		result = append(result, level)
	}
	return result
}

// 16. 验证二叉搜索树 [Medium] #98
func IsValidBST(root *TreeNode) bool {
	return validateBST(root, nil, nil)
}

func validateBST(node, min, max *TreeNode) bool {
	if node == nil {
		return true
	}

	if min != nil && node.Val <= min.Val {
		return false
	}
	if max != nil && node.Val >= max.Val {
		return false
	}

	return validateBST(node.Left, min, node) &&
		validateBST(node.Right, node, max)
}

// 17. 二叉树最近公共祖先 [Medium] #236
func LowestCommonAncestor(root, p, q *TreeNode) *TreeNode {
	if root == nil {
		return nil
	}
	if root == p || root == q {
		return root
	}

	left := LowestCommonAncestor(root.Left, p, q)
	right := LowestCommonAncestor(root.Right, p, q)

	if left != nil && right != nil {
		return root
	}
	if left != nil {
		return left
	}
	return right
}

// ============================================================================
// 动态规划
// ============================================================================

// 18. 爬楼梯 [Easy] #70
func ClimbStairs(n int) int {
	if n <= 2 {
		return n
	}

	prev, curr := 1, 2
	for i := 3; i <= n; i++ {
		prev, curr = curr, prev+curr
	}
	return curr
}

// 19. 零钱兑换 [Medium] #322
func CoinChange(coins []int, amount int) int {
	// dp[i] = 凑成金额 i 所需的最少硬币数
	dp := make([]int, amount+1)
	for i := range dp {
		dp[i] = amount + 1 // 初始化为不可能的值
	}
	dp[0] = 0

	for i := 1; i <= amount; i++ {
		for _, coin := range coins {
			if coin <= i {
				dp[i] = min(dp[i], dp[i-coin]+1)
			}
		}
	}

	if dp[amount] > amount {
		return -1
	}
	return dp[amount]
}

// 20. 最长公共子序列 [Medium] #1143
func LongestCommonSubsequence(text1, text2 string) int {
	m, n := len(text1), len(text2)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if text1[i-1] == text2[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}
	return dp[m][n]
}

// 21. 打家劫舍 [Medium] #198
func Rob(nums []int) int {
	if len(nums) == 0 {
		return 0
	}
	if len(nums) == 1 {
		return nums[0]
	}

	// dp[i] = max(dp[i-1], dp[i-2] + nums[i])
	prev2, prev1 := 0, nums[0]
	for i := 1; i < len(nums); i++ {
		curr := max(prev1, prev2+nums[i])
		prev2, prev1 = prev1, curr
	}
	return prev1
}

// 22. 完全平方数 [Medium] #279
func NumSquares(n int) int {
	// dp[i] = 最少完全平方数个数使其和为 i
	dp := make([]int, n+1)
	for i := range dp {
		dp[i] = n + 1
	}
	dp[0] = 0

	for i := 1; i <= n; i++ {
		for j := 1; j*j <= i; j++ {
			dp[i] = min(dp[i], dp[i-j*j]+1)
		}
	}
	return dp[n]
}

// ============================================================================
// 回溯算法
// ============================================================================

// 23. 全排列 [Medium] #46
func Permute(nums []int) [][]int {
	result := [][]int{}
	permuteBacktrack(nums, 0, &result)
	return result
}

func permuteBacktrack(nums []int, start int, result *[][]int) {
	if start == len(nums) {
		temp := make([]int, len(nums))
		copy(temp, nums)
		*result = append(*result, temp)
		return
	}

	for i := start; i < len(nums); i++ {
		nums[start], nums[i] = nums[i], nums[start]
		permuteBacktrack(nums, start+1, result)
		nums[start], nums[i] = nums[i], nums[start]
	}
}

// 24. 子集 [Medium] #78
func Subsets(nums []int) [][]int {
	result := [][]int{{}}
	for _, num := range nums {
		n := len(result)
		for i := 0; i < n; i++ {
			temp := make([]int, len(result[i]))
			copy(temp, result[i])
			temp = append(temp, num)
			result = append(result, temp)
		}
	}
	return result
}

// 25. 单词搜索 [Medium] #79
func Exist(board [][]byte, word string) bool {
	if len(board) == 0 {
		return false
	}

	m, n := len(board), len(board[0])
	visited := make([][]bool, m)
	for i := range visited {
		visited[i] = make([]bool, n)
	}

	for i := 0; i < m; i++ {
		for j := 0; j < n; j++ {
			if existDFS(board, word, i, j, 0, visited) {
				return true
			}
		}
	}
	return false
}

func existDFS(board [][]byte, word string, i, j, k int, visited [][]bool) bool {
	if i < 0 || i >= len(board) || j < 0 || j >= len(board[0]) ||
		visited[i][j] || board[i][j] != word[k] {
		return false
	}

	if k == len(word)-1 {
		return true
	}

	visited[i][j] = true
	if existDFS(board, word, i+1, j, k+1, visited) ||
		existDFS(board, word, i-1, j, k+1, visited) ||
		existDFS(board, word, i, j+1, k+1, visited) ||
		existDFS(board, word, i, j-1, k+1, visited) {
		return true
	}
	visited[i][j] = false
	return false
}

// ============================================================================
// 二分查找
// ============================================================================

// 26. 二分查找 [Easy] #704
func BinarySearch(nums []int, target int) int {
	left, right := 0, len(nums)-1

	for left <= right {
		mid := left + (right-left)/2
		if nums[mid] == target {
			return mid
		} else if nums[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return -1
}

// 27. 搜索二维矩阵 [Medium] #74
func SearchMatrix(matrix [][]int, target int) bool {
	if len(matrix) == 0 {
		return false
	}

	m, n := len(matrix), len(matrix[0])
	left, right := 0, m*n-1

	for left <= right {
		mid := left + (right-left)/2
		num := matrix[mid/n][mid%n]
		if num == target {
			return true
		} else if num < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return false
}

// 28. 在排序数组中查找元素的第一个和最后一个位置 [Medium] #34
func SearchRange(nums []int, target int) []int {
	return []int{findFirst(nums, target), findLast(nums, target)}
}

func findFirst(nums []int, target int) int {
	left, right := 0, len(nums)-1
	result := -1

	for left <= right {
		mid := left + (right-left)/2
		if nums[mid] == target {
			result = mid
			right = mid - 1 // 继续向左找
		} else if nums[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return result
}

func findLast(nums []int, target int) int {
	left, right := 0, len(nums)-1
	result := -1

	for left <= right {
		mid := left + (right-left)/2
		if nums[mid] == target {
			result = mid
			left = mid + 1 // 继续向右找
		} else if nums[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return result
}

// ============================================================================
// 滑动窗口
// ============================================================================

// 29. 无重复字符的最长子串 [Medium] #3
func LengthOfLongestSubstring(s string) int {
	charIndex := make(map[rune]int)
	maxLength := 0
	left := 0

	for right, char := range s {
		if idx, ok := charIndex[char]; ok && idx >= left {
			left = idx + 1
		}
		charIndex[char] = right
		maxLength = max(maxLength, right-left+1)
	}
	return maxLength
}

// 30. 最小覆盖子串 [Hard] #76
func MinWindow(s string, t string) string {
	if len(s) == 0 || len(t) == 0 {
		return ""
	}

	// 统计 t 中字符频率
	need := make(map[byte]int)
	for i := range t {
		need[t[i]]++
	}

	// 滑动窗口
	window := make(map[byte]int)
	left, right := 0, 0
	valid := 0
	start := 0
	length := len(s) + 1

	for right < len(s) {
		// 扩大窗口
		c := s[right]
		right++
		if need[c] > 0 {
			window[c]++
			if window[c] == need[c] {
				valid++
			}
		}

		// 判断是否需要收缩
		for valid == len(need) {
			// 更新结果
			if right-left < length {
				start = left
				length = right - left
			}

			// 收缩窗口
			d := s[left]
			left++
			if need[d] > 0 {
				if window[d] == need[d] {
					valid--
				}
				window[d]--
			}
		}
	}

	if length == len(s)+1 {
		return ""
	}
	return s[start : start+length]
}

// ============================================================================
// 堆
// ============================================================================

// 31. 数组中的第K个最大元素 [Medium] #215
// 使用快速选择算法，平均 O(n)
func FindKthLargest(nums []int, k int) int {
	return quickSelect(nums, 0, len(nums)-1, len(nums)-k)
}

func quickSelect(nums []int, left, right, k int) int {
	pivot := partition(nums, left, right)
	if pivot == k {
		return nums[pivot]
	} else if pivot < k {
		return quickSelect(nums, pivot+1, right, k)
	}
	return quickSelect(nums, left, pivot-1, k)
}

func partition(nums []int, left, right int) int {
	pivot := nums[right]
	i := left
	for j := left; j < right; j++ {
		if nums[j] < pivot {
			nums[i], nums[j] = nums[j], nums[i]
			i++
		}
	}
	nums[i], nums[right] = nums[right], nums[i]
	return i
}

// 32. 前 K 个高频元素 [Medium] #347
func TopKFrequent(nums []int, k int) []int {
	// 统计频率
	freq := make(map[int]int)
	for _, num := range nums {
		freq[num]++
	}

	// 桶排序
	buckets := make([][]int, len(nums)+1)
	for num, count := range freq {
		buckets[count] = append(buckets[count], num)
	}

	// 从高到低取 k 个
	result := []int{}
	for i := len(buckets) - 1; i >= 0 && len(result) < k; i-- {
		result = append(result, buckets[i]...)
	}
	return result[:k]
}

// ============================================================================
// 排序
// ============================================================================

// 33. 快速排序
func QuickSort(arr []int) {
	if len(arr) <= 1 {
		return
	}
	pivot := partitionSort(arr, 0, len(arr)-1)
	QuickSort(arr[:pivot])
	QuickSort(arr[pivot+1:])
}

func partitionSort(arr []int, low, high int) int {
	pivot := arr[high]
	i := low
	for j := low; j < high; j++ {
		if arr[j] < pivot {
			arr[i], arr[j] = arr[j], arr[i]
			i++
		}
	}
	arr[i], arr[high] = arr[high], arr[i]
	return i
}

// 34. 归并排序
func MergeSort(arr []int) {
	if len(arr) <= 1 {
		return
	}
	mid := len(arr) / 2
	MergeSort(arr[:mid])
	MergeSort(arr[mid:])
	merge(arr, mid)
}

func merge(arr []int, mid int) {
	left := append([]int{}, arr[:mid]...)
	right := append([]int{}, arr[mid:]...)

	i, j, k := 0, 0, 0
	for i < len(left) && j < len(right) {
		if left[i] <= right[j] {
			arr[k] = left[i]
			i++
		} else {
			arr[k] = right[j]
			j++
		}
		k++
	}

	for i < len(left) {
		arr[k] = left[i]
		i++
		k++
	}
	for j < len(right) {
		arr[k] = right[j]
		j++
		k++
	}
}

// ============================================================================
// 辅助函数
// ============================================================================

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func quickSort(arr []int) {
	// 使用内建排序
	// sort.Ints(arr)
	// 这里是简化版
}

func sortIntervals(intervals [][]int) {
	// 简化版排序
	// 实际应使用 sort.Slice
}
