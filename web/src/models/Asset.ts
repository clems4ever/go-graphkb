
export interface Asset {
    _id: string;
    type: string;
    key: string;
}

export type AssetWithSources = Asset & {sources: string[]}