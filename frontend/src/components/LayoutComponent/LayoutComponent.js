import React from 'react';
import { Box, CssBaseline } from '@mui/material';
import TopbarComponent from '../TopbarComponent/TopbarComponent';
import SidebarComponent from '../SidebarComponent/SidebarComponent';
import { Outlet } from 'react-router-dom';

const LayoutComponent = ({ children }) => {
    return (
        <Box sx={{ display: 'flex' }}>
            <CssBaseline />
            <TopbarComponent />
            <SidebarComponent />
            <Outlet />
        </Box>
    );
};

export default LayoutComponent;
