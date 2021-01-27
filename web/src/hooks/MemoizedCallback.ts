import React, { DependencyList } from "react";

export const useMemoizedCallback = <T extends (...args: any[]) => any>(callback: T, inputs: DependencyList) => {
    // Instance var to hold the actual callback.
    const callbackRef = React.useRef(callback);

    // The memoized callback that won't change and calls the changed callbackRef.
    const memoizedCallback = React.useCallback((...args) => {
        return callbackRef.current(...args);
    }, []);

    // The callback that is constantly updated according to the inputs.
    // eslint-disable-next-line
    const updatedCallback = React.useCallback(callback, inputs);

    // The effect updates the callbackRef depending on the inputs.
    React.useEffect(() => {
        callbackRef.current = updatedCallback;
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, inputs);

    // Return the memoized callback.
    return memoizedCallback;
};
