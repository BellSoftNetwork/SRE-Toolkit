package utils

import "strings"

// ExtractClusterName 클러스터 이름 추출
// arn:aws:eks:ap-northeast-2:{계정ID}:cluster/cluster1 형식에서 클러스터 이름만 추출
func ExtractClusterName(fullName string) string {
	parts := strings.Split(fullName, "/")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return fullName
}
