import axios from "axios";
import { Asset } from "../models/Asset";
import { QueryAssetsSources, QueryRelationsSources, QueryResultSet, QueryResultSetWithSources } from "../models/QueryResultSet";
import { DatabaseDetails } from "../models/DatabaseDetails";

export async function getSources() {
    const res = await axios.get<string[]>(`/api/sources`);

    if (res.status !== 200) {
        throw new Error(`Status code (${res.status}`);
    }

    return res.data;
}

export async function getDatabaseDetails() {
    const res = await axios.get<DatabaseDetails>(`/api/database`);

    if (res.status !== 200) {
        throw new Error(`Status code (${res.status}`);
    }

    return res.data;
}

export async function getSourceGraph(sourceNames: string[]) {
    const res = await axios.get(`/api/schema?sources=${sourceNames.join(",")}`);

    if (res.status !== 200) {
        throw new Error(`Status code (${res.status}`);
    }

    return res.data;
}

export interface SearchAssetResponse {
    assets: Asset[];
    total_hits: number;
}

export async function searchAssets(query: string, from: number = 0, size: number = 20) {
    const encodedQuery = encodeURI(query);
    const res = await axios.get<SearchAssetResponse>(`/search/assets?q=${encodedQuery}&from=${from}&size=${size}`);

    if (res.status !== 200) {
        throw new Error(`Status code (${res.status}`);
    }
    return res.data;
}

export async function postQuery(query: string) {
    const res = await axios.post<QueryResultSetWithSources>("/api/query", {
        q: query,
        include_sources: true,
    }, { validateStatus: s => s === 200 || s === 500 || s === 400 });

    if (res.status !== 200) {
        throw new Error(`${res.data} (${res.status})`);
    }
    return res.data;
}

export async function postAssetsSources(ids: string[]) {
    const res = await axios.post<QueryAssetsSources>("/api/query/assets/sources", {
        ids: ids,
    }, { validateStatus: s => s === 200 || s === 500 || s === 400 });

    if (res.status !== 200) {
        throw new Error(`${res.data} (${res.status})`);
    }
    return res.data;
}

export async function postRelationsSources(ids: string[]) {
    const res = await axios.post<QueryRelationsSources>("/api/query/relations/sources", {
        ids: ids,
    }, { validateStatus: s => s === 200 || s === 500 || s === 400 });

    if (res.status !== 200) {
        throw new Error(`${res.data} (${res.status})`);
    }
    return res.data;
}