import React, { useState, useEffect } from 'react';
import Alert from '@mui/material/Alert';  
import { Box, Button, Container } from '@mui/material';
import { Link } from 'react-router-dom';
import axios from 'axios';

const apiURL = process.env.REACT_APP_BACKEND_URL;

const InfoComponent = () => {
    const [message, setMessage] = useState('');
    const [severity, setSeverity] = useState('info');
    const [loading, setLoading] = useState(true);
    const location = window.location.pathname;

    const downloadFiles = async () => {
        try {
            const response = await axios.get(`${apiURL}/download_data`);
            const data = response.data;
            console.log('Data:', data);
            setMessage(data.message);
            setSeverity('success');
            localStorage.setItem('hasFiles', true);
        } catch (error) {
            console.error('Error fetching files:', error);
            setMessage(`Error fetching files: ${error}`);
            setSeverity('error');
            localStorage.setItem('hasFiles', false);
        } finally {
            setLoading(false);
        }
    };

    const uploadData = async () => {
        try {
            const response = await axios.get(`${apiURL}/load_data`);
            const data = response.data;
            setMessage(data.message);
            setSeverity('success');
        } catch (error) {
            console.error('Error upload data:', error);
            setMessage(`Error upload data: ${error}`);
            setSeverity('error');
        } finally {
            setLoading(false);
        }
    };

    const processPredictionsAssessments = async () => {
        try {
            const response = await axios.post(`${apiURL}/process_data_prediction_assessments`);
            const data = response.data;
            setMessage(data.message);
            setSeverity('success');
        } catch (error) {
            console.error('Error processing predictions assessments:', error);
            setMessage(`Error processing predictions assessments: ${error}`);
            setSeverity('error');
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if (location === '/info/download') {
            setMessage('Downloading files...');
            downloadFiles();
        }
        else if (location === '/info/upload') {
            setMessage('Uploading data...');
            uploadData();
        }
        else if (location === '/info/process_prediction_assessments') {
            setMessage('Processing prediction assessments data...');
            processPredictionsAssessments();
        }
    }, [location]);

    if (loading) {
        return (
            <Container component="main" maxWidth="md" sx={{ mt: 10, p: 2, bgcolor: 'background.paper', boxShadow: 3, borderRadius: 2, minWidth:'80%', minHeight: '100%' }}>
                <Alert severity={severity}>{message}</Alert>
            </Container>
        );
    }

    return (
        <Container component="main" maxWidth="md" sx={{ mt: 10, p: 2, bgcolor: 'background.paper', boxShadow: 3, borderRadius: 2, minWidth:'80%', minHeight: '100%' }}>
            <Alert severity={severity}>{message}</Alert>
            <Box>
                <br/>
                <Button 
                    variant="contained" 
                    size='small'
                    color={severity === 'success' ? 'primary' : 'secondary'}
                    component={Link}
                    to="/"
                >
                  Go to Home
                </Button>
            </Box>
        </Container>
    );
}

export default InfoComponent;
