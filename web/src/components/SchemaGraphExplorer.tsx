import React from "react";
import { useSchemaGraph } from "../hooks/SchemaGraph";
import D3Graph from "./D3Graph";

export interface Props {
    sources: string[];

    backgroundColor: string;
}

interface D3Node {
    id: string;
    label: string;
}

interface D3Link {
    id: string;
    source: string;
    target: string;
    label: string;
}

export default function SchemaGraphExplorer (props: Props) {
    const [graph, ,] = useSchemaGraph(props.sources);

    const nodes = graph ? graph.vertices
        .map(v => ({ id: v, label: v, } as D3Node))
        : [];

    const links = graph ? graph.edges
        .map(e => ({
            id: `${e.from_type}-${e.relation_type}-${e.to_type}`,
            source: e.from_type,
            target: e.to_type,
            label: e.relation_type
        } as D3Link))
        : [];

    return (
        <D3Graph nodes={nodes} edges={links}
            firstSimulationTick={50}
            forceLinkDistance={300}
            forceCollideRadius={60}
            forceManyBodyStrength={-200} />
    )
}
