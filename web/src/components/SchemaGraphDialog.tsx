import React, { useState } from "react";
import { Dialog, useTheme, makeStyles, List, ListItem, ListItemIcon, Checkbox, ListItemText, Switch } from "@material-ui/core";
import SchemaGraphExplorer from "./SchemaGraphExplorer";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faTimes } from "@fortawesome/free-solid-svg-icons";

export interface Props {
    sources: string[];

    open: boolean;
    onClose: () => void;
}

export default function SchemaGraphDialog (props: Props) {
    const theme = useTheme();
    const styles = useStyles();
    const [selectedSources, setSelectedSources] = useState<string[]>([]);

    const handleSourceClick = (source: string) => {
        if (selectedSources.indexOf(source) === -1) {
            setSelectedSources(selectedSources.concat([source]));
        } else {
            setSelectedSources(selectedSources.filter(s => s !== source));
        }
    }

    // useEffect(() => { setSelectedSources(sources) }, [sources]);

    return (
        <Dialog open={props.open}
            onClose={props.onClose}
            fullScreen
            className={styles.dialog}
            PaperProps={{ className: styles.dialogPaper }}>
            <div className={styles.schemaExplorerContainer}>
            <div className={styles.schemaGraphExplorer}>
                    <SchemaGraphExplorer
                        backgroundColor={theme.palette.background.default}
                        sources={selectedSources} />
                </div>
                <div className={styles.leftControl}>
                    <div className={styles.leftControlChild}>
                        <SourcesList
                            sources={props.sources.sort()}
                            selected={selectedSources}
                            className={styles.sourcesList}
                            onSourceClick={handleSourceClick} />
                    </div>
                </div>
                <FontAwesomeIcon
                    icon={faTimes}
                    className={styles.closeIcon}
                    size="2x" onClick={props.onClose}
                    style={{ width: 32 }} />
            </div>
        </Dialog>
    )
}

const useStyles = makeStyles(theme => ({
    dialog: {
        padding: theme.spacing(4),
    },
    dialogPaper: {
        borderRadius: "10px",
        overflow: "hidden",
    },
    schemaExplorerContainer: {
        height: '100%',
    },
    closeIcon: {
        position: "absolute",
        right: theme.spacing(2),
        top: theme.spacing(2),
        cursor: "pointer",
        color: "grey",
        opacity: 0.5,
        '&:hover': {
            opacity: 0.7
        }
    },
    leftControl: {
        padding: theme.spacing(),
        display: "inline-block",
        height: `calc(100% - ${2 * theme.spacing()}px)`,
        zIndex: 10000,
    },
    leftControlChild: {
        display: "inline-block",
        border: '1px solid grey',
        backgroundColor: "rgba(23, 23, 23, 0)",
        opacity: 0.7,
        '&:hover': {
            backgroundColor: "rgba(23, 23, 23, 1)",
        },
        '-ms-overflow-style': 'none',  /* IE and Edge */
        scrollbarWidth: 'none',
        '&::-webkit-scrollbar': {
            display: "none",
        },
        overflow: "auto",
        height: "100%",
    },
    sourcesList: {
    },
    schemaGraphExplorer: {
        position: "absolute",
        top: 0,
        left: 0,
        width: "100%",
        height: "100%",
    }
}));

interface SourcesListProps {
    sources: string[];
    selected: string[];

    className?: string;

    onSourceClick: (source: string) => void;
}

function SourcesList(props: SourcesListProps) {
    const handleToggle = (source: string) => {
        return () => props.onSourceClick(source);
    }

    const items = props.sources.map((s, i) => {
        return (
            <ListItem key={`item-${i}`} dense={true} onClick={handleToggle(s)}>
                <ListItemIcon>
                    <Checkbox
                        color="default"
                        edge="start"
                        checked={props.selected.indexOf(s) !== -1}
                        tabIndex={-1}
                        disableRipple
                        inputProps={{ 'aria-labelledby': `sources-list-item-${i}` }}
                    />
                </ListItemIcon>
                <ListItemText id={`sources-list-item-${i}`} primary={s} />
            </ListItem>
        )
    });

    return (
        <List className={props.className}>
            {items}
        </List >
    )
} 