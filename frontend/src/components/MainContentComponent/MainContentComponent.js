import React from 'react';
import { Toolbar, Box, Typography } from '@mui/material';

const MainContentComponent = ({ children }) => {
    return (
        <Box
            component="main"
            sx={{ flexGrow: 1, bgcolor: 'background.default', p: 3 }}
        >
            <Toolbar />
            <Typography paragraph>
                {children}
            </Typography>
        </Box>
    );
};

export default MainContentComponent;
