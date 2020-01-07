

export interface SourceEdge {
    relation_type: string;
    from_type: string;
    to_type: string;
}

export interface SourceGraph {
    vertices: string[];
    edges: SourceEdge[];
}