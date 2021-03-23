
export interface Relation {
    _id: string;
    type: string;
    from_id: string;
    to_id: string;
}

export type RelationWithSources = Relation & {sources: string[]}