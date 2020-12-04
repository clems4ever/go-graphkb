import React, { useEffect, useState, Fragment, useCallback } from "react";
import { QueryResultSet } from "../models/QueryResultSet";
import { Relation } from "../models/Relation";
import { Asset } from "../models/Asset";
import D3Graph from "./D3Graph";
import classnames from "classnames";
import { makeStyles, useTheme } from "@material-ui/core";


export interface Props {
    result: QueryResultSet | undefined;
    backgroundColor: string;

    onAssetDoubleClick: (asset: Asset) => void;
}

interface D3Node {
    id: string;
    label: string;
    asset: Asset;
}

interface D3Link {
    id: string;
    source: string;
    target: string;
    label: string;
    relation: Relation;
}

function uniqBy<T>(a: T[], key: (v: T) => string): T[] {
    var seen: { [k: string]: boolean } = {};
    return a.filter(function (item) {
        var k = key(item);
        return seen.hasOwnProperty(k) ? false : (seen[k] = true);
    })
}

function relationKey(r: Relation): string {
    return `${r.from_id}-${r.type}-${r.to_id}`;
}

export default function GraphExplorer (props: Props) {
    const [nodes, setNodes] = useState([] as D3Node[]);
    const [edges, setEdges] = useState([] as D3Link[]);
    const [assetHovered, setAssetHovered] = useState(undefined as Asset | undefined);
    const styles = useStyles();
    const maxEdges = 50;
    const theme = useTheme();

    const {
        result,
        onAssetDoubleClick
    } = props;

    const handleNodeDoubleClick = (d: D3Node) => onAssetDoubleClick(d.asset);
    const handleNodeMouseOver = useCallback((d: D3Node) => setAssetHovered(d.asset), [setAssetHovered]);
    const handleNodeMouseOut = useCallback((d: D3Node) => setAssetHovered(undefined), [setAssetHovered]);

    useEffect(() => {
        let assets: Asset[] = [];
        let relations: Relation[] = [];
        if (result && result.items) {
            for (const i in result.items) {
                const row = result.items[i]
                for (const j in row) {
                    const isAsset = result.columns[j].type === "asset";
                    const isRelation = result.columns[j].type === "relation";
                    if (isAsset) {
                        assets.push(row[j] as Asset);
                    } else if (isRelation) {
                        relations.push(row[j] as Relation);
                    }
                }
            }
            assets = uniqBy(assets, a => a._id);
            relations = uniqBy(relations, relationKey);
        }

        const selectedAssets = assets.slice(0, maxEdges);
        const selectedRelations: Relation[] = [];

        function isAssetExisting(assetID: string): boolean {
            return selectedAssets.filter(v => v._id === assetID).length === 1;
        }

        for (let x in relations.slice(0, maxEdges)) {
            const rel = relations[x];
            if (isAssetExisting(rel.from_id) && isAssetExisting(rel.to_id)) {
                selectedRelations.push(rel);
            }
        }

        const d3nodes = selectedAssets.map(a => ({ id: a._id, label: a.key, asset: a } as D3Node));
        const d3edges = selectedRelations.map(r => ({ id: relationKey(r), label: r.type, source: r.from_id, target: r.to_id, relation: r } as D3Link));
        setNodes(d3nodes);
        setEdges(d3edges);
    }, [result, setNodes, setEdges]);

    return (
        <Fragment>
            <div className={styles.elementDetailsContainer}>
                <div className={classnames(styles.elementDetails, !assetHovered ? "hidden" : "")}>
                    <b>type:</b> {assetHovered ? assetHovered.type : ""}<br /><b>value:</b> {assetHovered ? assetHovered.key : ""}
                </div>
            </div>
            <D3Graph
                nodes={nodes} edges={edges}
                backgroundColor={theme.palette.background.default}
                colorGroupResolver={d => d.asset.type}
                onNodeDoubleClick={handleNodeDoubleClick}
                onNodeMouseOut={handleNodeMouseOut}
                onNodeMouseOver={handleNodeMouseOver} />
        </Fragment>
    )
}

const useStyles = makeStyles(theme => ({
    elementDetailsContainer: {
        position: "relative",
        backgroundColor: theme.palette.background.paper,
    },
    elementDetails: {
        position: "absolute",
        backgroundColor: "black",
        color: "white",
        opacity: 0.4,
        padding: theme.spacing(1),
        top: theme.spacing(),
        right: theme.spacing(),
        borderRadius: "5px",
    },
}));