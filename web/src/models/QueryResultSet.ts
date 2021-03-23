import { Asset, AssetWithSources } from "./Asset";
import { Relation, RelationWithSources } from "./Relation";

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

export type TypedDocWithSources = AssetWithSources | RelationWithSources | string;

export type RowResponseWithSources = TypedDocWithSources[];

export interface QueryResultSetWithSources {
    items: RowResponseWithSources[];
    columns: ColumnType[];
    execution_time_ms: number;
}

export interface QueryAssetsSources {
    results: {[id: string]: string[]}
}

export interface QueryRelationsSources {
    results: {[id: string]: string[]}
}