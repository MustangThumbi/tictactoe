package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/MustangThumbi/tictactoe/genproto/tictactoe"

)

type server struct {
	pb.UnimplementedTictactoeServer
	mu    sync.Mutex
	games map[string]*Game
}

type Game struct {
	Board  [3][3]string
	Status string
}

// CreateGame initializes a new game with a unique ID
func (s *server) CreateGame(ctx context.Context, req *pb.CreateGameRequest) (*pb.CreateGameResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	gameID := uuid.New().String()
	s.games[gameID] = &Game{
		Status: "ongoing",
	}

	log.Printf("🟩 Created new game: %s", gameID)
	return &pb.CreateGameResponse{GameId: gameID}, nil
}

// MakeMove processes a player's move and updates the game state
func (s *server) MakeMove(ctx context.Context, req *pb.MakeMoveRequest) (*pb.MakeMoveResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	game, exists := s.games[req.GameId]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "game not found")
	}

	if game.Status != "ongoing" {
		return &pb.MakeMoveResponse{Status: game.Status}, nil
	}

	row, col := req.Row, req.Col
	player := req.Player

	if row < 0 || row > 2 || col < 0 || col > 2 || game.Board[row][col] != "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid move")
	}

	game.Board[row][col] = player
	game.Status = checkWinner(game.Board)

	boardFlat := flattenBoard(game.Board)

	log.Printf("🟩 Move by %s in game %s at [%d,%d]", player, req.GameId, row, col)
	return &pb.MakeMoveResponse{
		Status: game.Status,
		Board:  boardFlat,
	}, nil
}

// GetGameState returns the current state of a game
func (s *server) GetGameState(ctx context.Context, req *pb.GetGameStateRequest) (*pb.GetGameStateResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	game, exists := s.games[req.GameId]
	if !exists {
		return nil, status.Errorf(codes.NotFound, "game not found")
	}

	boardFlat := flattenBoard(game.Board)

	return &pb.GetGameStateResponse{
		Status: game.Status,
		Board:  boardFlat,
	}, nil
}

// checkWinner determines the game status based on the board
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

// flattenBoard converts a 3x3 board to a 9-element string slice
func flattenBoard(board [3][3]string) []string {
	var flat []string
	for _, row := range board {
		for _, cell := range row {
			flat = append(flat, cell)
		}
	}
	return flat
}

// runGRPCServer starts the gRPC server
func runGRPCServer(srv *server) {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf(" Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterTictactoeServer(grpcServer, srv)

	log.Println(" gRPC server running on :50051")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf(" Failed to serve gRPC: %v", err)
	}
}

// runHTTPGateway starts the REST gateway server with CORS
func runHTTPGateway() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := pb.RegisterTictactoeHandlerFromEndpoint(ctx, mux, "localhost:50051", opts)
	if err != nil {
		log.Fatalf(" Failed to register gRPC-Gateway: %v", err)
	}

	handler := cors.AllowAll().Handler(mux)

	log.Println("🚀 REST gateway running on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatalf(" Failed to serve HTTP: %v", err)
	}
}

func main() {
	srv := &server{
		games: make(map[string]*Game),
	}

	go runGRPCServer(srv)
	runHTTPGateway()
}
