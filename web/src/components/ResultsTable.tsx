import React, { useState, useEffect, memo } from "react";
import { QueryResultSet, TypedDoc, ColumnType } from "../models/QueryResultSet";
import MaterialTable from "material-table"
import { Asset } from "../models/Asset";
import { Relation } from "../models/Relation";
import { makeStyles } from "@material-ui/core";

export interface Props {
    results: QueryResultSet | undefined;
    isLoading: boolean;
}

function computeColumns(columns: ColumnType[]) {
    return columns.map((v, i) => ({ title: `${v.name} (${v.type})`, field: `col-${i}`, export: true }));
}

function cellToValue(row: TypedDoc[], colIdx: number, columns: ColumnType[]): string {
    const v = row[colIdx];
    if (columns[colIdx].type === "property") {
        return v as string;
    } else if (columns[colIdx].type === "asset") {
        const d = v as Asset;
        return d.key;
    } else if (columns[colIdx].type === "relation") {
        const d = v as Relation;
        return d.type;
    }
    return "unknown";
}

function columnToValue(results: QueryResultSet, rowIdx: number): { [k: string]: string } {
    const values = {} as { [k: string]: string };
    results.items[rowIdx].forEach((v, i) => {
        const x = cellToValue(results.items[rowIdx], i, results.columns);
        values[`col-${i}`] = x;
    });
    return values;
}

function computeValues(results: QueryResultSet) {
    if (results.items.length === 0) {
        return [];
    }
    return results.items.map((v, i) => columnToValue(results, i));
}

const ResultsTable = memo(function (props: Props) {
    const [columns, setColumns] = useState<{ title: string, field: string }[]>([]);
    const [data, setData] = useState<{ [k: string]: string }[]>([]);

    useEffect(() => {
        const cols: { title: string, field: string }[] = props.results
            ? computeColumns(props.results.columns)
            : [];
        setColumns(cols);

        const d = props.results
            ? computeValues(props.results)
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
