package service

// State определяет тип для состояния
type State int

// Определение возможных состояний с использованием iota
const (
	CONNECTED State = iota
	AUTHORIZATE
	SELECT_ACTION
	GET_DATA
	CHOSE_CREATE_DATA
	CREATE_DATA
)

type DataType int

const (
	PASSWORD DataType = iota
	TEXT
	CARD
)
