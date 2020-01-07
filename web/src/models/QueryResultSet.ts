import { Asset } from "./Asset";
import { Relation } from "./Relation";

export type TypedDoc = Asset | Relation | string;

export type RowResponse = TypedDoc[];

export interface ColumnType {
    name: string
    type: "asset" | "relation" | "property";
}

export interface QueryResultSet {
    items: RowResponse[];
    columns: ColumnType[];
    execution_time_ms: number;
}