import React, { useState, useEffect, memo } from "react";
import { ColumnType, QueryResultSetWithSources, TypedDocWithSources } from "../models/QueryResultSet";
import MaterialTable from "material-table"
import { AssetWithSources } from "../models/Asset";
import { RelationWithSources } from "../models/Relation";
import { makeStyles, Theme, Tooltip, Typography, useTheme } from "@material-ui/core";

export interface Props {
    results: QueryResultSetWithSources | undefined;
    isLoading: boolean;
}

function computeColumns(columns: ColumnType[]) {
    return columns.map((v, i) => ({ title: `${v.name} (${v.type})`, field: `col-${i}`, export: true }));
}

function cellToValue(row: TypedDocWithSources[], colIdx: number, columns: ColumnType[], theme: Theme): string | JSX.Element {
    const v = row[colIdx];
    if (columns[colIdx].type === "property") {
        return v as string;
    } else if (columns[colIdx].type === "asset") {
        const d = v as AssetWithSources;
        const key = (d.key === '') ? '(empty)' : d.key;
        return (
            <div>
                <span style={{color: "yellow"}}>{d.type}</span><br/>
                <span>{key}</span><br />
                <Tooltip title={<ul style={{listStyleType: "none", marginLeft: 0, paddingLeft: 0}}>{d.sources.map((x, i) => <li key={`li-${i}`}>{x}</li>)}</ul>}>
                <Typography style={{
                    fontSize: theme.typography.fontSize * 0.8,
                    color: "#a1ff8d",
                    maxWidth: 200,
                    whiteSpace: "nowrap",
                    overflow: "hidden",
                    textOverflow: "ellipsis"}
                }>
                    sources: {d.sources.join(', ')}
                </Typography>
                </Tooltip>
            </div>
        );
    } else if (columns[colIdx].type === "relation") {
        const d = v as RelationWithSources;
        return (
            <div>
                <span>{d.type}</span><br />
                <Tooltip title={<ul style={{listStyleType: "none", marginLeft: 0, paddingLeft: 0}}>{d.sources.map((x, i) => <li key={`li-${i}`}>{x}</li>)}</ul>}>
                    <Typography style={{
                        fontSize: theme.typography.fontSize * 0.8,
                        color: "#a1ff8d",
                        textOverflow: "ellipsis"
                    }}>
                        sources: {d.sources.join(',')}
                    </Typography>
                </Tooltip>
            </div>
        );
    }
    return "unknown";
}

function columnToValue(results: QueryResultSetWithSources, rowIdx: number, theme: Theme): { [k: string]: string | JSX.Element } {
    const values = {} as { [k: string]: string | JSX.Element };
    results.items[rowIdx].forEach((v, i) => {
        const x = cellToValue(results.items[rowIdx], i, results.columns, theme);
        values[`col-${i}`] = x;
    });
    return values;
}

function computeValues(results: QueryResultSetWithSources, theme: Theme) {
    if (results.items.length === 0) {
        return [];
    }
    return results.items.map((v, i) => columnToValue(results, i, theme));
}

const ResultsTable = memo(function (props: Props) {
    const theme = useTheme();
    const [columns, setColumns] = useState<{ title: string, field: string }[]>([]);
    const [data, setData] = useState<{ [k: string]: string | JSX.Element }[]>([]);

    useEffect(() => {
        const cols: { title: string, field: string }[] = props.results
            ? computeColumns(props.results.columns)
            : [];
        setColumns(cols);

        const d = props.results
            ? computeValues(props.results, theme)
            : [];
        setData(d);
    }, [props.results]);

    const classes = useStyles();

    return (
        <div className={classes.table}>
            <MaterialTable
                columns={columns}
                data={data}
                isLoading={props.isLoading}
                options={{
                    exportButton: true,
                    exportAllData: true,
                    exportFileName: "go-graphkb",
                    pageSize: 10,
                    pageSizeOptions: [10, 30, 50],
                    emptyRowsWhenPaging: false,
                    maxBodyHeight: "100%",
                }}
                title="Results of last query"
            />
        </div>
    )
});

const useStyles = makeStyles(theme => ({
    table: {
        overflow: "auto",
        maxHeight: "100%",
        height: "100%",
    }
}));

// ResultsTable.whyDidYouRender = true;

export default ResultsTable;
