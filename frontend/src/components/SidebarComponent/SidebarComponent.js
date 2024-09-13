import React from 'react';
import { Drawer, List, ListItemButton, ListItemIcon, ListItemText, Toolbar } from '@mui/material';
import DownloadIcon from '@mui/icons-material/Download';
import UploadIcon from '@mui/icons-material/Upload';
import AutorenewIcon from '@mui/icons-material/Autorenew';
import { Link } from 'react-router-dom';


const drawerWidth = 240;

const SidebarComponent = () => {

    return (
        <Drawer
            variant="permanent"
            sx={{
                width: drawerWidth,
                flexShrink: 0,
                [`& .MuiDrawer-paper`]: { width: drawerWidth, boxSizing: 'border-box' },
            }}
        >
            <Toolbar />
            <List>
                    <ListItemButton key="Download Data" component={Link} to="/info/download">
                        <ListItemIcon>
                            <DownloadIcon />
                        </ListItemIcon>
                        <ListItemText primary="Download Data"/>
                    </ListItemButton>
                    <ListItemButton key="Upload Data" component={Link} to="/info/upload">
                        <ListItemIcon>
                            <UploadIcon />
                        </ListItemIcon>
                        <ListItemText primary="Upload Data" />
                    </ListItemButton>
                    <ListItemButton key="Process Data Assessments" component={Link} to="/info/process_prediction_assessments">
                        <ListItemIcon>
                            <AutorenewIcon />
                        </ListItemIcon>
                        <ListItemText primary="Process Data Assessments" />
                    </ListItemButton>
                    {/* <ListItemButton key="Process Data VLE" component={Link} to="/info/process_prediction_vle">
                        <ListItemIcon>
                            <AutorenewIcon />
                        </ListItemIcon>
                        <ListItemText primary="Process Data VLE" />
                    </ListItemButton> */}
            </List>
        </Drawer>
    );
};

export default SidebarComponent;
