package main

import "strings"

func nodeID(node n8nNode) string {
	if node.ID != "" {
		return node.ID
	}
	return node.Name
}

func shortNodeType(value string) string {
	if value == "" {
		return ""
	}
	parts := strings.Split(value, ".")
	return parts[len(parts)-1]
}

func millisToUnix(value int64) int64 {
	if value <= 0 {
		return 0
	}
	if value > 100000000000 {
		return value / 1000
	}
	return value
}

func errorMessage(err n8nError) string {
	if err.Message != "" && err.Description != "" {
		return err.Message + ": " + err.Description
	}
	if err.Message != "" {
		return err.Message
	}
	if err.Description != "" {
		return err.Description
	}
	if len(err.Messages) > 0 {
		return strings.Join(err.Messages, "; ")
	}
	if err.Name != "" {
		return err.Name
	}
	return ""
}
