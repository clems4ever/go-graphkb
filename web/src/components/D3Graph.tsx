import React, { useRef, useEffect, useState } from "react";
import * as d3 from "d3";
import { makeStyles } from "@material-ui/core";
import Measure, { BoundingRect } from "react-measure";
import { useMemoizedCallback } from "../hooks/MemoizedCallback";

interface NodeLike {
    id: string;
    label: string;
}

interface EdgeLike {
    id: string;
    source: string;
    target: string;
    label: string;
}

export interface Props<Node, Edge> {
    nodes: Node[];
    edges: Edge[];

    nodeRadius?: number;

    forceLinkDistance?: number;
    forceCollideRadius?: number;
    forceManyBodyStrength?: number;
    firstSimulationTick?: number;

    backgroundColor?: string;

    colorGroupResolver?: (node: Node) => string;

    onNodeDoubleClick?: (node: Node) => void;
    onNodeMouseOver?: (node: Node) => void;
    onNodeMouseOut?: (node: Node) => void;
}

interface D3Node<NodeLike> {
    id: string;
    label: string;
    x: number;
    y: number;
    data: NodeLike;
}

interface D3Link<EdgeLike> {
    id: string;
    source: string;
    target: string;
    label: string;
    data: EdgeLike;
}

const AssetTypeToColor: { [type: string]: string } = {};

function pickOrAssignColor(type: string) {
    if (type in AssetTypeToColor) {
        return AssetTypeToColor[type];
    }
    const color = d3.interpolateWarm(Math.random());
    AssetTypeToColor[type] = color;
    return color;
}

let uniqueID = 0;

function linkPath(d: any) {
    if (d.selfLink) {
        var x1 = d.source.x,
            y1 = d.source.y,
            x2 = d.targetX,
            y2 = d.targetY,

            // Defaults for normal edge.
            drx = 40,
            dry = 40,
            xRotation = -45, // degrees
            largeArc = 1, // 1 or 0
            sweep = 1; // 1 or 0

        // For whatever reason the arc collapses to a point if the beginning
        // and ending points of the arc are the same, so kludge it.
        x2 = x2 + 1;
        y2 = y2 + 1;
        return "M" + x1 + "," + y1 + "A" + drx + "," + dry + " " + xRotation + "," + largeArc + "," + sweep + " " + x2 + "," + y2;
    }
    return 'M ' + d.source.x + ' ' + d.source.y + ' L ' + d.targetX + ' ' + d.targetY;
}

const D3Graph = <Node extends NodeLike, Edge extends EdgeLike>(props: Props<Node, Edge>) => {
    const containerRef = useRef<SVGSVGElement>(null);
    const arrowheadLength = 13;
    const classes = useStyles();
    const [bounds, setBounds] = useState(undefined as BoundingRect | undefined);
    const hoverDisabled = useRef<boolean>(false);

    const {
        nodeRadius,
        backgroundColor,
        nodes,
        edges,
        onNodeDoubleClick,
        onNodeMouseOut,
        onNodeMouseOver,
        forceCollideRadius,
        forceLinkDistance,
        forceManyBodyStrength,
        colorGroupResolver,
        firstSimulationTick,
    } = props;

    const onNodeMouseOverCallback = useMemoizedCallback((d: Node) => {
        if (onNodeMouseOver) onNodeMouseOver(d);
    }, [onNodeMouseOver]);

    const onNodeMouseOutCallback = useMemoizedCallback((d: Node) => {
        if (onNodeMouseOut) onNodeMouseOut(d);
    }, [onNodeMouseOut]);

    const onNodeDoubleClickCallback = useMemoizedCallback((d: Node) => {
        if (onNodeDoubleClick) onNodeDoubleClick(d);
    }, [onNodeDoubleClick]);

    const colorGroupResolverCallback = useMemoizedCallback((d: Node) => {
        return (colorGroupResolver) ? colorGroupResolver(d) : "";
    }, [colorGroupResolver]);

    useEffect(() => { uniqueID += 1; }, []);

    useEffect(() => {
        const svg = d3.select(containerRef.current);
        svg
            .append("rect")
            .attr("height", "100%")
            .attr("width", "100%")
            .attr("fill", backgroundColor ? backgroundColor : "none");

        svg.append("defs").append("marker")
            .attr("id", "arrow")
            .attr("viewBox", "0 -5 10 10")
            .attr("refX", 0)
            .attr("refY", 0)
            .attr("markerWidth", arrowheadLength)
            .attr("markerHeight", arrowheadLength)
            .attr("orient", "auto")
            .attr("fill", "white")
            .append("path")
            .attr("d", "M0,-5L10,0L0,5");

        svg.append("g").attr("class", "graph");
    }, [backgroundColor]);

    useEffect(() => {
        if (!bounds) {
            return;
        }
        const svg = d3.select(containerRef.current);
        const graph = svg.select(".graph");
        const zoom = d3.zoom()
            .scaleExtent([0.5, 2])
            .translateExtent([[-5000, -5000], [5000 + bounds.width, 5000 + bounds.height]])
            .on("zoom", function zoomed(e) {
                graph.attr("transform", e.transform);
            }
        );

        svg.call(zoom as any).on("dblclick.zoom", null);
    }, [bounds]);

    useEffect(() => {
        if (!bounds) {
            return;
        }

        const nRadius = nodeRadius ? nodeRadius : 40;
        const d3nodes = nodes
            .map(n => ({ id: n.id, label: n.label, data: n } as D3Node<Node>));
        const d3links = edges
            .map(e => ({ id: e.id, source: e.source, target: e.target, label: e.label } as D3Link<Edge>));

        const svg = d3.select(containerRef.current);
        const graph = svg.select(".graph");

        graph.selectAll(".link-group").remove();
        graph.selectAll(".node-group").remove();

        const linkGroup = graph
            .selectAll(".link-group")
            .data(d3links, (v: any) => v.id);

        const newLinkGroups = linkGroup
            .enter()
            .append("g")
            .attr("class", "link-group");

        newLinkGroups
            .append("path")
            .style("stroke", "white")
            .style("opacity", 0.6)
            .style("fill", "none")
            .attr("class", "link")
            .attr("marker-end", "url(#arrow)");

        newLinkGroups
            .append('path')
            .attr('fill-opacity', 0)
            .attr('stroke-opacity', 0)
            .attr('id', (d, i) => `graph${uniqueID}-edgepath${i}`)
            .attr('class', 'edgepath')
            .style("pointer-events", "none");

        const linkText = newLinkGroups
            .append('text')
            .style("pointer-events", "none")
            .attr('id', (d, i) => 'edgelabel' + i)
            .attr('font-size', 10)
            .attr('fill', 'white')
            .attr('class', 'edgelabel');

        linkText
            .append('textPath')
            .attr('xlink:href', (d, i) => `#graph${uniqueID}-edgepath${i}`)
            .style("text-anchor", "middle")
            .style("pointer-events", "none")
            .attr("startOffset", "50%")
            .attr('class', 'edgetextpath')
            .text(d => d.label);


        const nodeGroup = graph
            .selectAll(".node-group")
            .data(d3nodes, (v: any) => v.id);

        const newNodeGroups = nodeGroup
            .enter()
            .append("g")
            .attr("class", "node-group")
            .attr("z-index", "10")
            .on("mouseover", function (ev, d) {
                ev.preventDefault();
                d3.select(this).select("circle").attr("stroke", "white");
                d3.select(this).select("text").attr("stroke", "white");
                if (!hoverDisabled.current) {
                    onNodeMouseOverCallback(d.data);
                }
            })
            .on("mouseout", function (d) {
                d3.select(this).select("circle").attr("stroke", "none");
                d3.select(this).select("text").attr("stroke", "none");
                if (!hoverDisabled.current) {
                    onNodeMouseOutCallback(d.data);
                }
            })
            .on("dblclick", function (d) {
                onNodeDoubleClickCallback(d.data);
            });

        newNodeGroups
            .append("circle")
            .attr("r", nRadius)
            .style("fill", (d) => pickOrAssignColor(colorGroupResolverCallback(d.data)))
            .style("overflow", "hidden");

        const nodeText = newNodeGroups
            .append("text")
            .attr("text-anchor", "middle")
            .attr("fill", "white")
            .style("font-size", "0.7em");

        nodeText.append("tspan").text(d => d.label).each(function () {
            var self = d3.select(this),
                textLength = self.node()!.getComputedTextLength(),
                text = self.text();
            while (textLength > (80 - 2 * 5) && text.length > 0) {
                text = text.slice(0, -1);
                self.text(text + '...');
                textLength = self.node()!.getComputedTextLength();
            }
        })

        const simulation = d3.forceSimulation(d3nodes)
            .force("collide", d3.forceCollide(forceCollideRadius ? forceCollideRadius : 50))
            .force("charge", d3.forceManyBody().strength(forceManyBodyStrength ? forceManyBodyStrength : -100))
            .force("link", d3.forceLink(d3links).id((d: any) => d.id)
                .distance(forceLinkDistance ? forceLinkDistance : 200))
            .force("center", d3.forceCenter(bounds.width / 2, bounds.height / 2))
            .on("tick", ticked);

        simulation.tick(firstSimulationTick ? firstSimulationTick : 100);

        function ticked() {
            graph
                .selectAll(".link-group")
                .select(".link")
                .each(function (d: any) {
                    var x1 = d.source.x,
                        y1 = d.source.y,
                        x2 = d.target.x,
                        y2 = d.target.y,
                        angle = Math.atan2(y2 - y1, x2 - x1);
                    d.targetX = x2 - Math.cos(angle) * (nRadius + arrowheadLength);
                    d.targetY = y2 - Math.sin(angle) * (nRadius + arrowheadLength);
                    d.selfLink = d.target.id === d.source.id;
                });

            graph
                .selectAll(".link-group")
                .select(".edgepath")
                .attr('d', linkPath)
                .filter((d: any) => d.selfLink)
                .attr("transform", (d: any) => `rotate(135, ${d.source.x}, ${d.source.y})`)


            graph
                .selectAll(".link-group")
                .select(".edgelabel")
                .attr('transform', function (this: any, d: any) {
                    if (d.target.x < d.source.x) {
                        var bbox = this.getBBox();

                        const rx = bbox.x + bbox.width / 2;
                        const ry = bbox.y + bbox.height / 2;
                        return 'rotate(180 ' + rx + ' ' + ry + ')';
                    }
                    return 'rotate(0)';
                });

            graph
                .selectAll(".link-group")
                .select(".link")
                .attr('d', linkPath)
                .filter((d: any) => d.selfLink)
                .attr("transform", (d: any) => `rotate(135, ${d.source.x}, ${d.source.y})`)

            graph
                .selectAll(".node-group")
                .attr("transform", (d: any) => `translate(${d.x}, ${d.y})`);
        }

        const dragDrop: any = d3.drag()
            .on('start', (ev: any, node: any) => {
                hoverDisabled.current = true;
                if (!ev.active) simulation.alphaTarget(0.1).restart();
                node.fx = node.x;
                node.fy = node.y;
            })
            .on('drag', (ev: any, node: any) => {
                node.fx = ev.x;
                node.fy = ev.y;
            })
            .on('end', (ev: any, node: any) => {
                hoverDisabled.current = false;
                if (!ev.active) simulation.alphaTarget(0)
                node.fx = null;
                node.fy = null;
            })
        graph.selectAll(".node-group").call(dragDrop);
    }, [
        bounds,
        nodes,
        edges,
        forceCollideRadius,
        forceLinkDistance,
        forceManyBodyStrength,
        nodeRadius,
        onNodeDoubleClickCallback,
        onNodeMouseOutCallback,
        onNodeMouseOverCallback,
        colorGroupResolverCallback,
        firstSimulationTick
    ]);

    return (
        <Measure bounds onResize={(e) => setBounds(e.bounds)}>
            {({ measureRef }) => (
                <div className={classes.svgContainer} ref={measureRef}>
                    <svg
                        style={{ display: "block" }}
                        ref={containerRef}
                        className={classes.svg}>
                    </svg>
                </div>)}
        </Measure>
    )
}

const useStyles = makeStyles(theme => ({
    svg: {
        width: "100%",
        height: "100%",
    },
    svgContainer: {
        width: "100%",
        height: "100%",
    }
}));

export default D3Graph;