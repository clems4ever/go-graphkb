import React from "react";
import { Dialog, makeStyles, Typography } from "@material-ui/core";

export interface Props {
    open: boolean;

    assetsCount: number;
    relationsCount: number;

    onClose: () => void;
}

export default function DatabaseDialog(props: Props) {
    const classes = useStyles();
    return (
        <Dialog
            open={props.open}
            onClose={props.onClose}>
            <p className={classes.content}>
                <Typography variant="h4">Database details</Typography>
                <p>
                    Number of assets: {props.assetsCount}<br />
                    Number of relations: {props.relationsCount}
                </p>
            </p>
        </Dialog>
    )
}

const useStyles = makeStyles(theme => ({
    content: {
        margin: theme.spacing(2),
    }
}))