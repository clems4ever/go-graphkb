import React, { ChangeEvent, FocusEvent, KeyboardEvent, RefObject } from "react";
import { Paper, InputBase, makeStyles, createStyles } from "@material-ui/core";
import SearchIcon from '@material-ui/icons/Search';
import classnames from "classnames";

export interface Props {
    value: string;
    className?: string;
    inputRef?: RefObject<HTMLInputElement>;

    onChange: (event: ChangeEvent<HTMLInputElement>) => void;
    onFocus?: (event: FocusEvent<HTMLInputElement>) => void;
    onBlur?: (event: FocusEvent<HTMLInputElement>) => void;
    onKeyDown?: (event: KeyboardEvent<HTMLInputElement>) => void;
    onEnterKeyDown?: () => void;
}

export default function SearchField (props: Props) {
    const classes = useStyles();

    const handleKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
        if (event.key === "Enter") {
            if (props.onEnterKeyDown) {
                props.onEnterKeyDown();
            }
        }
        if (props.onKeyDown) {
            props.onKeyDown(event);
        }
    }

    return (
        <Paper className={classnames(classes.root, props.className)} classes={{
            rounded: classes.rounded,
        }}>
            <InputBase
                value={props.value}
                ref={props.inputRef}
                className={classes.input}
                placeholder="Search assets"
                onChange={props.onChange}
                onFocus={props.onFocus}
                onBlur={props.onBlur}
                onKeyDown={handleKeyDown}
            />
            <SearchIcon className={classes.iconButton} />
        </Paper>
    )
}

const useStyles = makeStyles((theme) =>
    createStyles({
        root: {
            padding: '2px 4px',
            display: 'flex',
            alignItems: 'center',
            width: "100%",
        },
        input: {
            marginLeft: theme.spacing(2),
            flex: 1,
        },
        rounded: {
            borderRadius: '30px',
        },
        iconButton: {
            padding: 10,
        },
    }),
);