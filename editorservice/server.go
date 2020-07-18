package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/golang/glog"
	empty "github.com/golang/protobuf/ptypes/empty"
	pgx "github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	pb "github.com/nick96/cycad/editorservice/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type editorServiceServer struct {
	pb.UnimplementedEditorServiceServer
	db *pgxpool.Pool
}

func parseLinks(content string) []*pb.Link {
	return []*pb.Link{}
}

func getBacklinksByNodeID(db *pgxpool.Pool, ctx context.Context, id uuid.UUID) ([]*pb.Backlink, error) {
	backlinkRows, err := db.Query(ctx, `
SELECT node.name
FROM editor.nodes as node, editor.links as link
WHERE link.toNode = $1
`, id)
	if err != nil && err != pgx.ErrNoRows {
		glog.Errorf("Failed to retrieve backlinks for node %v: %v", id, err)
		return nil, status.Errorf(codes.Unavailable, "Failed to retrieve backlinks for node %v", id)
	}

	var backlinks []*pb.Backlink
	for backlinkRows.Next() {
		var name string
		if err := backlinkRows.Scan(&name); err != nil {
			glog.Errorf("Failed to scan backlink name out of query: %v", err)
			return nil, status.Errorf(codes.Unavailable, "Failed to retrieve backlinks for node %v", id)
		}
		backlinks = append(backlinks, &pb.Backlink{
			Name: name,
		})
	}
	return backlinks, nil
}

func getLinksByNodeID(db *pgxpool.Pool, ctx context.Context, id uuid.UUID) ([]*pb.Link, error) {
	linkRows, err := db.Query(ctx, `
SELECT node.name
FROM editor.nodes as node, editor.links as link
WHERE link.fromNode = $1
`, id)
	if err != nil && err != pgx.ErrNoRows {
		glog.Errorf("Failed to retrieve links for node %v: %v", id, err)
		return nil, status.Errorf(codes.Unavailable, "Failed to retrieve links for node %v", id)
	}
	var links []*pb.Link
	var name string
	for linkRows.Next() {
		if err := linkRows.Scan(&name); err != nil {
			glog.Errorf("Failed to scan link name out of query: %v", err)
			return nil, status.Errorf(codes.Unavailable, "Failed to retrieve links for node %v", id)
		}
		links = append(links, &pb.Link{Name: name})
	}
	return links, nil
}

func (s *editorServiceServer) AddNode(ctx context.Context, req *pb.AddNodeRequest) (*pb.AddNodeResponse, error) {
	glog.Infof("Received AddNode request: %v", req)

	if strings.TrimSpace(req.Name) == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Empty names are not valid")
	}

	if _, err := s.db.Exec(
		ctx,
		"INSERT INTO editor.nodes(name, content) values ($1, $2) ON CONFLICT DO NOTHING",
		req.Name, req.Content,
	); err != nil {
		glog.Errorf("Failed to create node with name %s: %v", req.Name, err)
		return nil, status.Errorf(codes.Unavailable, "Failed to create node with name %s", req.Name)
	}

	var fromNodeID uuid.UUID
	err := s.db.QueryRow(ctx, "SELECT id FROM editor.nodes WHERE name = $1", req.Name).Scan(&fromNodeID)
	if err != nil {
		// Something is not right! We just ensured that this node existed.
		glog.Errorf("Failed to retrieve node with name: %s: %v", req.Name, err)
		return nil, status.Errorf(codes.Unavailable, "Failed to create node with name %s", req.Name)
	}

	links := parseLinks(req.Content)
	for _, link := range links {
		var toNodeID uuid.UUID
		if err := s.db.QueryRow(
			ctx,
			`
INSERT INTO editor.nodes (name, content)
VALUES($1, $2)
ON CONFLICT (name) DO NOTHING
RETURNING id`,
			link.Name, "").Scan(&toNodeID); err != nil {
			glog.Errorf("Failed to upsert node with name %s: %v", link.Name, err)
			return nil, status.Errorf(codes.Unavailable, "Failed on link with name %s", link.Name)
		}
		if _, err := s.db.Exec(
			ctx,
			`
INSERT INTO editor.links (fromNode, toNode, pos) VALUES ($1, $2, $3)
`,
			fromNodeID, toNodeID, link.Offset); err != nil {
			glog.Errorf("Failed to insert link between %v and %v: %v", fromNodeID, toNodeID, err)
			return nil, status.Errorf(codes.Unavailable, "Failed on link with name %s", link.Name)
		}
	}

	backlinks, err := getBacklinksByNodeID(s.db, ctx, fromNodeID)
	if err != nil {
		return nil, err
	}
	return &pb.AddNodeResponse{Name: req.Name, Content: req.Content, Backlinks: backlinks, Links: links}, nil
}

func (s *editorServiceServer) GetNode(ctx context.Context, req *pb.GetNodeRequest) (*pb.GetNodeResponse, error) {
	glog.Infof("Received GetNode request: %v", req)

	id, err := uuid.FromString(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "ID '%s' is not a valid UUID", req.Id)
	}

	var name, content string
	err = s.db.QueryRow(
		ctx,
		"SELECT name, content FROM editor.nodes WHERE id = $1",
		id,
	).Scan(&name, &content)
	if err != nil && err == pgx.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "No node with ID %s was found", id)
	} else if err != nil {
		glog.Errorf("Failed to get node by ID %v: %v", id, err)
		// There was a problem. We've logged it at let the user know
		// that it's not them, it's us. Hopefully it works if the user
		// trys again.
		return nil, status.Errorf(codes.Unavailable, "Failed to get node with ID %s", id)
	}

	backlinks, err := getBacklinksByNodeID(s.db, ctx, id)
	if err != nil {
		return nil, err
	}

	links, err := getLinksByNodeID(s.db, ctx, id)
	if err != nil {
		return nil, err
	}

	return &pb.GetNodeResponse{Name: "name", Content: "content", Backlinks: backlinks, Links: links}, nil
}

func (s *editorServiceServer) GetAllNodes(req *empty.Empty, srv pb.EditorService_GetAllNodesServer) error {
	glog.Info("Received get all logs request")
	ctx := srv.Context()
	rows, err := s.db.Query(ctx, "SELECT id, name FROM editor.nodes")
	if err != nil {
		glog.Errorf("Failed to gets nodes: %v", err)
		return status.Errorf(codes.Unavailable, "Failed to get nodes")
	}

	var id uuid.UUID
	var name string
	for rows.Next() {
		if err := rows.Scan(&id, &name); err != nil {
			glog.Errorf("Failed to scan row in all nodes query: %v", err)
			return status.Errorf(codes.Unavailable, "Failed to get nodes")
		}

		links, err := getLinksByNodeID(s.db, ctx, id)
		if err != nil {
			return err
		}

		backlinks, err := getBacklinksByNodeID(s.db, ctx, id)
		if err != nil {
			return err
		}

		// Just send back the node without the content because we're not (that) crazy...
		if err := srv.Send(&pb.GetNodeResponse{Name: name, Backlinks: backlinks, Links: links}); err != nil {
			glog.Errorf("Failed to send response: %v", err)
			return status.Errorf(codes.Unavailable, "Failed to send response")
		}
	}

	glog.Info("Finished streaming all nodes")
	return nil
}

func newServer(db *pgxpool.Pool) *editorServiceServer {
	svr := &editorServiceServer{db: db}
	return svr
}

func checkMigrationVersion(db *pgxpool.Pool) error {
	var ver int32
	if err := db.QueryRow(
		context.Background(),
		"SELECT version FROM schema_version",
	).Scan(&ver); err != nil {
		return fmt.Errorf("failed to retrieve schema version number: %w", err)
	}

	if ver != expectedMigrationVersion {
		return fmt.Errorf("expected migration version %d but found %d", expectedMigrationVersion, ver)
	}

	return nil
}
