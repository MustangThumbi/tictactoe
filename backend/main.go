package main

import (
	"context"
	"log"
	"net"
	"sync"

	pb "github.com/MustangThumbi/tictactoe/genproto/tictactoe"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedTictactoeServer
	mu    sync.Mutex
	games map[string]*Game
}

// Game struct holds the game state
type Game struct {
	Board  [3][3]string
	Status string
}

func (s *server) CreateGame(ctx context.Context, req *pb.CreateGameRequest) (*pb.CreateGameResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	gameID := "game123" // In practice, generate a UUID
	s.games[gameID] = &Game{
		Status: "ongoing",
	}

	return &pb.CreateGameResponse{GameId: gameID}, nil
}

func (s *server) MakeMove(ctx context.Context, req *pb.MakeMoveRequest) (*pb.MakeMoveResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	game, exists := s.games[req.GameId]
	if !exists {
		return nil, grpc.Errorf(404, "game not found")
	}

	if game.Status != "ongoing" {
		return &pb.MakeMoveResponse{Status: game.Status}, nil
	}

	row, col := req.Row, req.Col
	player := req.Player

	if row < 0 || row > 2 || col < 0 || col > 2 || game.Board[row][col] != "" {
		return nil, grpc.Errorf(400, "invalid move")
	}

	game.Board[row][col] = player
	game.Status = checkWinner(game.Board)

	boardFlat := []string{}
	for _, r := range game.Board {
		for _, c := range r {
			boardFlat = append(boardFlat, c)
		}
	}

	return &pb.MakeMoveResponse{
		Status: game.Status,
		Board:  boardFlat,
	}, nil
}

func (s *server) GetGameState(ctx context.Context, req *pb.GetGameStateRequest) (*pb.GetGameStateResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	game, exists := s.games[req.GameId]
	if !exists {
		return nil, grpc.Errorf(404, "game not found")
	}

	boardFlat := []string{}
	for _, r := range game.Board {
		for _, c := range r {
			boardFlat = append(boardFlat, c)
		}
	}

	return &pb.GetGameStateResponse{
		Status: game.Status,
		Board:  boardFlat,
	}, nil
}

// Check for winner or draw
func checkWinner(board [3][3]string) string {
	lines := [][3][2]int{
		{{0, 0}, {0, 1}, {0, 2}},
		{{1, 0}, {1, 1}, {1, 2}},
		{{2, 0}, {2, 1}, {2, 2}},
		{{0, 0}, {1, 0}, {2, 0}},
		{{0, 1}, {1, 1}, {2, 1}},
		{{0, 2}, {1, 2}, {2, 2}},
		{{0, 0}, {1, 1}, {2, 2}},
		{{0, 2}, {1, 1}, {2, 0}},
	}

	for _, line := range lines {
		a, b, c := line[0], line[1], line[2]
		if board[a[0]][a[1]] != "" &&
			board[a[0]][a[1]] == board[b[0]][b[1]] &&
			board[a[0]][a[1]] == board[c[0]][c[1]] {
			return board[a[0]][a[1]] + "_wins"
		}
	}

	for _, row := range board {
		for _, cell := range row {
			if cell == "" {
				return "ongoing"
			}
		}
	}

	return "draw"
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterTictactoeServer(grpcServer, &server{
		games: make(map[string]*Game),
	})

	log.Println("gRPC server running on port 50051...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
