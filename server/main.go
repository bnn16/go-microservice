package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "grpc/ms/pb"

	_ "github.com/mattn/go-sqlite3"
)

const (
	port       = ":50051"
	dbFileName = "items.db"
)

type server struct {
	pb.UnimplementedMyServiceServer
	db *sql.DB
}

// function to add items to the database (CREATE)
func (s *server) AddItem(ctx context.Context, req *pb.ItemRequest) (*pb.ItemResponse, error) {

	//insert the item into the database
	result, err := s.db.Exec("INSERT INTO items(name) VALUES(?)", req.Name)
	//check for errors, due to sqlite instance not running?
	if err != nil {
		return nil, fmt.Errorf("failed to add item: %v", err)
	}

	//get the id of the last inserted item
	id, _ := result.LastInsertId()

	return &pb.ItemResponse{
		Id:   id,
		Name: req.Name,
	}, nil
}

func main() {
	//listen on port 50051
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	//open the database file
	db, err := sql.Open("sqlite3", dbFileName)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	//check if the database is alive
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	//create the table if it doesn't exist
	sqlStmt := `
		CREATE TABLE IF NOT EXISTS items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT
		);
	`
	if _, err := db.Exec(sqlStmt); err != nil {
		log.Fatalf("failed to create table: %v", err)
	}

	//create the gRPC server/api
	s := grpc.NewServer()
	pb.RegisterMyServiceServer(s, &server{db: db})

	log.Printf("Server listening on port %s", port)
	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
