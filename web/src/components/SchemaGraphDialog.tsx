import React, { useState, useEffect } from "react";
import { Dialog, useTheme, makeStyles, List, ListItem, ListItemIcon, Checkbox, ListItemText, Switch } from "@material-ui/core";
import SchemaGraphExplorer from "./SchemaGraphExplorer";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faTimes } from "@fortawesome/free-solid-svg-icons";

export interface Props {
    sources: string[];

    open: boolean;
    onClose: () => void;
}

export default function (props: Props) {
    const theme = useTheme();
    const styles = useStyles();
    const [selectedSources, setSelectedSources] = useState<string[]>([]);
    const [hideObservations, setHideObservations] = useState(true);
    const { sources } = props;

    const handleSourceClick = (source: string) => {
        if (selectedSources.indexOf(source) === -1) {
            setSelectedSources(selectedSources.concat([source]));
        } else {
            setSelectedSources(selectedSources.filter(s => s !== source));
        }
    }

    useEffect(() => { setSelectedSources(sources) }, [sources]);

    return (
        <Dialog open={props.open}
            onClose={props.onClose}
            fullScreen
            className={styles.dialog}
            PaperProps={{ className: styles.dialogPaper }}>
            <div className={styles.schemaExplorerContainer}>
                <div className={styles.leftControl}>
                    <SourcesList
                        sources={props.sources.sort()}
                        selected={selectedSources}
                        showObservation={!hideObservations}
                        onShowObservationsClick={() => setHideObservations(!hideObservations)}
                        className={styles.sourcesList}
                        onSourceClick={handleSourceClick} />
                </div>
                <FontAwesomeIcon
                    icon={faTimes}
                    className={styles.closeIcon}
                    size="2x" onClick={props.onClose}
                    style={{ width: 32 }} />
                <SchemaGraphExplorer
                    backgroundColor={theme.palette.background.default}
                    sources={selectedSources}
                    hideObservations={hideObservations} />
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
        height: "100%",
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
        position: "absolute",
        top: theme.spacing(),
        left: theme.spacing(),
        borderRadius: 5,
        backgroundColor: "rgba(23, 23, 23, 0)",
        opacity: 0.7,
        '&:hover': {
            backgroundColor: "rgba(23, 23, 23, 1)",
        }
    },
    sourcesList: {
    }
}));

interface SourcesListProps {
    sources: string[];
    selected: string[];
    showObservation: boolean;

    className?: string;

    onSourceClick: (source: string) => void;
    onShowObservationsClick: () => void;
}

function SourcesList(props: SourcesListProps) {
    const handleToggle = (source: string) => {
        return () => props.onSourceClick(source);
    }

    const handleShowObservationsToggle = () => {
        props.onShowObservationsClick();
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
            <ListItem dense={true} onClick={handleShowObservationsToggle}>
                <ListItemIcon>
                    <Switch
                        size="small"
                        color="default"
                        edge="start"
                        checked={props.showObservation}
                        tabIndex={-1}
                        disableRipple
                        inputProps={{ 'aria-labelledby': `show-source-observations-item` }}
                    />
                </ListItemIcon>
                <ListItemText id={`show-source-observations-item`} primary={"Show observations"} />
            </ListItem>
            {items}
        </List >
    )
} 