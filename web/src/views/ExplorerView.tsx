import React, { useState, useEffect, useCallback } from 'react';
import { makeStyles, Grid, Snackbar, Paper, useTheme } from '@material-ui/core';
import GraphExplorer from '../components/GraphExplorer';
import QueryField from '../components/QueryField';
import { postQuery, getSources } from "../services/SourceGraph";
import { QueryResultSetWithSources } from '../models/QueryResultSet';
import ResultsTable from '../components/ResultsTable';
import { Asset } from '../models/Asset';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faProjectDiagram, faDatabase, IconDefinition } from '@fortawesome/free-solid-svg-icons'
import SchemaGraphDialog from '../components/SchemaGraphDialog';
import { SizeProp } from '@fortawesome/fontawesome-svg-core';
import DatabaseDialog from '../components/DatabaseDialog';
import SearchField from '../components/SearchField';
import MuiAlert from '@material-ui/lab/Alert';
import { useQueryParam, StringParam, withDefault } from 'use-query-params';

function Alert(props: any) {
  return <MuiAlert elevation={6} variant="filled" {...props} />;
}

const QUERY = `MATCH (n0)-[r]-(n1) RETURN n0, r, n1 LIMIT 20`;

const ExplorerView = () => {
    const theme = useTheme();
    const styles = useStyles();
    const [sources, setSources] = useState(undefined as string[] | undefined);
    const [query, setQuery] = useState(QUERY);
    const [submittedQuery, setSubmittedQuery] = useQueryParam("q", withDefault(StringParam, QUERY));
    const [queryResult, setQueryResult] = useState(undefined as QueryResultSetWithSources | undefined);
    const [isQueryLoading, setIsQueryLoading] = useState(false);
    const [error, setError] = useState(undefined as undefined | Error);
    const [schemaOpen, setSchemaOpen] = useState(false);
    const [databaseDialogOpen, setDatabaseDialogOpen] = useState(false);
    const [searchValue, setSearchValue] = useState("");
    const [searchFocus, setSearchFocus] = useState(false);

    const handleQuerySubmit = useCallback(async (q: string) => {
        setSubmittedQuery(q);
    }, [setSubmittedQuery]);

    useEffect(() => {
        (async function() {
            setIsQueryLoading(true);
            try {
                const res = await postQuery(submittedQuery);
                setQueryResult(res);
            } catch (err) {
                console.error(err);
                setError(err);
            }
            setIsQueryLoading(false);
        })();
        setQuery(submittedQuery);
    }, [submittedQuery]);

    const getSourcesCallback = useCallback(async () => {
        try {
            setSources(await getSources());
        } catch (err) {
            console.error(err);
            setError(new Error("Unable to fetch sources: " + err.message));
        }
    }, []);

    useEffect(() => { handleQuerySubmit(QUERY) }, [handleQuerySubmit]);

    useEffect(() => { getSourcesCallback() }, [getSourcesCallback]);

    const handleAssetDoubleClick = (asset: Asset) => {
        const newQuery = `MATCH (a:${asset.type})-[r]-(n) WHERE a.value = '${asset.key}' RETURN a, r, n`;
        setQuery(newQuery);
        handleQuerySubmit(newQuery);
    }

    const handleSchemaIconClick = () => {
        setSchemaOpen(true);
    }

    const handleDatabaseIconClick = () => {
        setDatabaseDialogOpen(true);
    }

    const handleSearchRequested = () => {
        const newQuery = `MATCH (a) WHERE a.value CONTAINS '${searchValue}' RETURN a`;
        setQuery(newQuery);
        handleQuerySubmit(newQuery);
    }

    return (
        <div>
            <Snackbar open={error !== undefined}
                onClose={() => setError(undefined)}
                anchorOrigin={{
                    vertical: 'top',
                    horizontal: 'right',
                }}>
                    <Alert onClose={() => setError(undefined)} severity="error">
                        {error ? error.message : ""}
                    </Alert>
            </Snackbar>
            <div className={styles.container}>
                <DatabaseDialog
                    open={databaseDialogOpen}
                    onClose={() => setDatabaseDialogOpen(false)} />

                <SchemaGraphDialog
                    open={schemaOpen}
                    onClose={() => setSchemaOpen(false)}
                    sources={sources ? sources : []} />

                <Grid container className={styles.gridContainer}>
                    <Grid item xs={8}>
                        <div className={styles.graphExplorerContainer}>
                            <div className={styles.graphExplorerContainerInner}>
                                <div className={styles.searchContainer}>
                                    <div className={styles.searchContainerCentered}>
                                        <SearchField
                                            className={(searchFocus) ? styles.searchField : styles.searchFieldTransparent}
                                            value={searchValue}
                                            onFocus={() => setSearchFocus(true)}
                                            onBlur={() => setSearchFocus(false)}
                                            onChange={e => setSearchValue(e.target.value)}
                                            onEnterKeyDown={handleSearchRequested} />
                                    </div>
                                </div>

                                <ButtonGroup buttons={[
                                    { icon: faProjectDiagram, onClick: handleSchemaIconClick },
                                    { icon: faDatabase, onClick: handleDatabaseIconClick },
                                ]} />
                                <GraphExplorer
                                    backgroundColor={theme.palette.background.default}
                                    result={queryResult}
                                    onAssetDoubleClick={handleAssetDoubleClick} />
                                <div className={styles.queryFieldContainer}>
                                    <div className={styles.queryFieldContainerInner}>
                                        <div className={styles.resultsSummary}>
                                            {queryResult ? <span>{queryResult.items.length} results founds in {queryResult.execution_time_ms}ms</span> : null}
                                        </div>
                                        <QueryField
                                            query={query}
                                            onChange={setQuery}
                                            onSubmit={() => handleQuerySubmit(query)} />
                                    </div>
                                </div>
                            </div>
                        </div>
                    </Grid>
                    <Grid item xs={4}>
                        <div className={styles.resultsContainer}>
                            <div className={styles.resultsContainerInner}>
                                <ResultsTable results={queryResult} isLoading={isQueryLoading} />
                            </div>
                        </div>
                    </Grid>
                </Grid>
            </div>
        </div >
    );
}

const useStyles = makeStyles(theme => ({
    container: {
        position: 'absolute',
        width: "100%",
        height: "100%",
    },
    gridContainer: {
        height: '100%',
        flexWrap: "nowrap",
    },
    graphExplorerContainer: {
        height: "100%",
        display: "flex",
    },
    graphExplorerContainerInner: {
        position: "relative",
        padding: theme.spacing(),
        flex: 1,
    },
    resultsContainer: {
        overflowY: "auto",
        overflowX: "hidden",
        height: "100%",
    },
    resultsContainerInner: {
        margin: theme.spacing(),
    },
    searchContainer: {
        position: 'absolute',
        width: '500px',
        top: theme.spacing(2),
        left: '50%',
        zIndex: 100,
    },
    searchContainerCentered: {
        position: "relative",
        left: "-50%",
    },
    searchField: {
        opacity: 0.95,
    },
    searchFieldTransparent: {
        opacity: 0.1,
        '&:hover': {
            opacity: 0.95,
        }
    },
    queryFieldContainer: {
        position: 'absolute',
        bottom: '0px',
        left: '0px',
        width: "100%",
    },
    queryFieldContainerInner: {
        margin: theme.spacing(),
        padding: theme.spacing(2),
    },
    resultsSummary: {
        paddingBottom: theme.spacing(),
        color: "grey",
        fontSize: 0.9 * theme.typography.fontSize,
    },
    queryHint: {
        position: "absolute",
        right: theme.spacing(2),
    }
}));

interface Button {
    icon: IconDefinition;
    size?: SizeProp;
    onClick: () => void;
}

interface ButtonGroupProps {
    buttons: Button[];
}

function ButtonGroup(props: ButtonGroupProps) {
    const classes = makeStyles(theme => ({
        schemaIconContainer: {
            position: "relative",
            textAlign: "center",
        },
        schemaIconContainerInner: {
            position: "absolute",
            left: theme.spacing(),
            top: theme.spacing(),
        },
        schemaIcon: {
            padding: theme.spacing(2),
            marginBottom: theme.spacing(),
            backgroundColor: "grey",
            borderRadius: "10px",
            opacity: 0.3,
            '&:hover': {
                opacity: 0.5,
                cursor: "pointer",
            }
        },
    }))();

    const items = props.buttons.map((b, i) => {
        return (
            <Paper
                className={classes.schemaIcon}
                elevation={2}
                onClick={b.onClick}
                key={`button-${i}`}>
                <FontAwesomeIcon icon={b.icon} size={b.size} />
            </Paper>
        )
    });

    return (
        <div className={classes.schemaIconContainer}>
            <div className={classes.schemaIconContainerInner}>
                {items}
                {/* <Paper className={classes.schemaIcon} elevation={2} onClick={handleSchemaIconClick}>
                    <FontAwesomeIcon icon={faProjectDiagram} />
                </Paper>
                <Paper className={classes.schemaIcon} elevation={2} onClick={handleSchemaIconClick}>
                    <FontAwesomeIcon icon={faDatabase} size="lg" />
    </Paper> */}
            </div>
        </div>
    )
}



export default ExplorerView;
