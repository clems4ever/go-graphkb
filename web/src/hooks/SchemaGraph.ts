import { useState, useEffect, useCallback } from "react";
import { getSourceGraph } from "../services/SourceGraph";
import { SourceGraph } from "../models/SourceGraph";


export function useSchemaGraph(sourceNames: string[]): [SourceGraph | undefined, Error | undefined, () => void] {
    const [graph, setGraph] = useState<SourceGraph | undefined>();
    const [error, setError] = useState<Error | undefined>();

    const clearError = useCallback(() => setError(undefined), [setError]);

    const fetchSource = useCallback(async () => {
        try {
            setGraph(await getSourceGraph(sourceNames));
        } catch (err) {
            setError(err);
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [sourceNames]);

    useEffect(() => { fetchSource() }, [fetchSource]);

    return [graph, error, clearError];
}