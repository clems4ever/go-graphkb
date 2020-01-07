import React, { useRef, useState, FocusEvent, useCallback, KeyboardEvent, useEffect, ChangeEvent, MouseEvent } from "react";
import SearchField from "./SearchField";
import { Fade, Paper, makeStyles, Typography, Popper, List, ListItemText, ListItem, ClickAwayListener } from "@material-ui/core";
import { Asset } from "../models/Asset";
import { SearchAssetResponse, searchAssets } from "../services/SourceGraph";

let timer: NodeJS.Timeout | undefined;

export interface Props {
    onResultClick: (asset: Asset) => void;
}

export default function (props: Props) {
    const [value, setValue] = useState("");
    const [isSearchFocus, setIsSearchFocus] = useState(false);
    const [popoverOpen, setPopoverOpen] = useState(false);
    const [assets, setAssets] = useState(undefined as SearchAssetResponse | undefined);

    const searchContainerRef = useRef<HTMLDivElement>(null);
    const searchFieldRef = useRef<HTMLInputElement>(null);
    const searchSelectedResultRef = useRef<HTMLDivElement>(null);
    const styles = useStyles();

    const handleSearchChange = (event: ChangeEvent<HTMLInputElement>) => {
        const searchValue = event.target.value;
        setValue(event.target.value);

        if (searchValue === "") {
            setAssets(undefined);
            return;
        }
        if (timer) {
            clearTimeout(timer);
            timer = undefined;
        }
        timer = setTimeout(async () => {
            const res = await searchAssets(searchValue, 0, 10);
            setAssets(res);
        }, 250);
    }


    const handleFocus = (event: FocusEvent<HTMLInputElement>) => {
        setIsSearchFocus(true);
    }

    const handleAssetClick = (asset: Asset) => {
        setTimeout(() => setPopoverOpen(false), 150);
        props.onResultClick(asset);
    }

    const handleSearcKeyDown = useCallback((event: KeyboardEvent<HTMLInputElement>) => {
        // On arrow down we send the focus to the asset list
        if (event.keyCode === 40 && searchSelectedResultRef.current) {
            searchSelectedResultRef.current!.focus();
        }
    }, [searchSelectedResultRef]);

    useEffect(() => {
        setPopoverOpen(isSearchFocus && value !== "");
    }, [isSearchFocus, value]);

    return (
        <div ref={searchContainerRef} className={styles.root}>
            <SearchField
                value={value}
                inputRef={searchFieldRef}
                onChange={handleSearchChange}
                onFocus={handleFocus}
                onKeyDown={handleSearcKeyDown} />
            <Popper
                open={popoverOpen}
                anchorEl={searchContainerRef.current}
                transition>
                {({ TransitionProps }) => (
                    <Fade {...TransitionProps} timeout={350}>
                        <Paper elevation={1} className={styles.searchResultsContainer}>
                            {assets !== undefined
                                ? <AssetList
                                    assets={assets.assets}
                                    totalHits={assets.total_hits}
                                    onClickAway={() => setPopoverOpen(false)}
                                    onAssetClick={handleAssetClick} />
                                : null}
                        </Paper>
                    </Fade>)}
            </Popper>
        </div>
    )
}

const useStyles = makeStyles(theme => ({
    root: {
        display: "block",
    },
    searchResultsContainer: {
        width: "450px",
        marginTop: theme.spacing(),
    }
}));

interface AssetListProps {
    assets: Asset[];
    totalHits: number;

    onClickAway: () => void;
    onAssetClick: (asset: Asset) => void;
}

function AssetList(props: AssetListProps) {
    const classes = makeStyles(theme => ({
        resultsFound: {
            padding: theme.spacing(),
            paddingLeft: theme.spacing(2),
            paddingRight: theme.spacing(2),
            textAlign: "right",
        }
    }))();

    const handleAssetClick = (asset: Asset) => {
        return (event: MouseEvent) => props.onAssetClick(asset);
    }

    const items = props.assets.map((it, i) => {
        return <ListItem
            button
            onClick={handleAssetClick(it)}
            key={`asset-${i}`}
            className={"active"}>
            <ListItemText primary={it.key} secondary={it.type} />
        </ListItem>
    });

    return (
        <ClickAwayListener onClickAway={props.onClickAway}>
            <div>
                <List dense>{items}</List>
                <Typography className={classes.resultsFound}>
                    {props.totalHits} results found
                </Typography>
            </div>
        </ClickAwayListener>
    )
}