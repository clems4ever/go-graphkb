import React, { useCallback, useEffect, useState } from "react";
import { Dialog, makeStyles, Snackbar, Typography } from "@material-ui/core";
import { getDatabaseDetails } from "../services/SourceGraph";
import Alert from "@material-ui/lab/Alert";
import { DatabaseDetails } from "../models/DatabaseDetails";

export interface Props {
    open: boolean;

    onClose: () => void;
}

export default function DatabaseDialog(props: Props) {
    const classes = useStyles();

    const [error, setError] = useState(undefined as undefined | Error);
    const [databaseDetails, setDatabaseDetails] = useState(undefined as DatabaseDetails | undefined);

    const getDatabaseDetailsCallback = useCallback(async () => {
        try {
            setDatabaseDetails(await getDatabaseDetails());
        } catch (err) {
            console.error(err);
            setError(new Error("Unable to fetch database details: " + err.message));
        }
    }, []);

    useEffect(() => { if (props.open) getDatabaseDetailsCallback() }, [getDatabaseDetailsCallback, props.open]);

    return (
        <>
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
        <Dialog
            open={props.open}
            onClose={props.onClose}>
            <p className={classes.content}>
                <Typography variant="h4">Database details</Typography>
                <p>
                    Number of assets: {databaseDetails?.assets_count}<br />
                    Number of relations: {databaseDetails?.relations_count}
                </p>
            </p>
        </Dialog>
        </>
    )
}

const useStyles = makeStyles(theme => ({
    content: {
        margin: theme.spacing(2),
    }
}))