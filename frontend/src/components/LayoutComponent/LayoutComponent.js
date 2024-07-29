import React from 'react';
import { Box, CssBaseline } from '@mui/material';
import TopbarComponent from '../TopbarComponent/TopbarComponent';
import SidebarComponent from '../SidebarComponent/SidebarComponent';
import MainContentComponent from '../MainContentComponent/MainContentComponent';

const LayoutComponent = ({ children }) => {
    return (
        <Box sx={{ display: 'flex' }}>
            <CssBaseline />
            <TopbarComponent />
            <SidebarComponent />
            <MainContentComponent>
                {children}
            </MainContentComponent>
        </Box>
    );
};

export default LayoutComponent;
