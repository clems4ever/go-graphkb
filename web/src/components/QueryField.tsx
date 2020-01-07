import React, { KeyboardEvent, useRef, useState } from "react";
import { Button, TextField, Icon, makeStyles, Tooltip } from "@material-ui/core";
import { Cursor } from "../models/Cursor";

export interface Props {
    query: string;

    onChange: (q: string) => void;
    onSubmit: () => void;
}

export default function (props: Props) {
    const styles = useStyles();
    const inputRef = useRef<HTMLInputElement>();
    const [cursor, setCursor] = useState({ line: 0, column: 0 } as Cursor);

    const splitQuery = props.query.split("\n");
    const rows = Math.max(2, splitQuery.length);

    const handleOnKeyDown = (e: KeyboardEvent) => {
        if (e.key === "Enter" && e.ctrlKey) {
            props.onSubmit();
        }
    }

    const updateSelection = () => {
        if (inputRef.current!.selectionStart === inputRef.current!.selectionEnd) {
            const idx = inputRef.current!.selectionStart!;
            let line: number = 0, column: number = 0;
            for (let i = 0; i < idx; i++) {
                column += 1;
                if (props.query[i] === "\n") {
                    column = 0;
                    line += 1;
                }
            }
            setCursor({ line, column });
        }
    }

    return (
        <div className={styles.root}>
            <div className={styles.container}>
                <div className={styles.rightControl}>
                    <Tooltip title="Ctrl+Enter">
                        <Button
                            variant="outlined"
                            onClick={() => props.onSubmit()}
                            className={styles.button}>
                            <Icon>send</Icon>
                        </Button>
                    </Tooltip>
                    <div className={styles.cursor}>col {cursor.column} : row {cursor.line + 1}</div>
                </div>
                <TextField multiline fullWidth
                    variant="outlined"
                    rows={rows}
                    value={props.query}
                    autoComplete="off"
                    autoCorrect="off"
                    inputRef={inputRef}
                    onClick={updateSelection}
                    onChange={e => {
                        props.onChange(e.target.value)
                    }}
                    onKeyDown={handleOnKeyDown}
                    onKeyUp={updateSelection}
                    InputProps={{ className: styles.inputBase }} />
            </div>
        </div>
    )
}

const useStyles = makeStyles(theme => ({
    rightControl: {
        margin: theme.spacing(),
        position: "absolute",
        right: 0,
        zIndex: 100,
    },
    button: {
    },
    cursor: {
        color: "white",
        textAlign: "center",
        fontSize: theme.typography.fontSize * 0.8,
        marginTop: theme.spacing(0.5),
        opacity: 0.5,
    },
    inputBase: {
        paddingRight: "100px",
    },
    root: {
        backgroundColor: "#424242",
    },
    container: {
        position: "relative",
    },
}));