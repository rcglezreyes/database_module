import React from 'react';
import { Drawer, List, ListItemButton, ListItemIcon, ListItemText, Toolbar } from '@mui/material';
import DownloadIcon from '@mui/icons-material/Download';
import UploadIcon from '@mui/icons-material/Upload';
import AutorenewIcon from '@mui/icons-material/Autorenew';

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
                    <ListItemButton key="Download Data">
                        <ListItemIcon>
                            <DownloadIcon />
                        </ListItemIcon>
                        <ListItemText primary="Download Data" />
                    </ListItemButton>
                    <ListItemButton key="Upload Data">
                        <ListItemIcon>
                            <UploadIcon />
                        </ListItemIcon>
                        <ListItemText primary="Upload Data" />
                    </ListItemButton>
                    <ListItemButton key="Process Data">
                        <ListItemIcon>
                            <AutorenewIcon />
                        </ListItemIcon>
                        <ListItemText primary="Process Data" />
                    </ListItemButton>
            </List>
        </Drawer>
    );
};

export default SidebarComponent;
