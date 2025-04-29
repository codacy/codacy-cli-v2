def complex_analysis(data, options=None):
    """A function with high complexity"""
    if not data:
        return None
    
    results = {
        'summary': {},
        'details': [],
        'warnings': [],
        'errors': []
    }
    
    # Process different types of data
    if isinstance(data, dict):
        for key, value in data.items():
            if isinstance(value, (int, float)):
                if value > 100:
                    results['summary'][key] = 'high'
                    results['details'].append({
                        'type': 'numeric',
                        'value': value,
                        'status': 'high'
                    })
                elif value > 50:
                    results['summary'][key] = 'medium'
                    results['details'].append({
                        'type': 'numeric',
                        'value': value,
                        'status': 'medium'
                    })
                else:
                    results['summary'][key] = 'low'
                    results['details'].append({
                        'type': 'numeric',
                        'value': value,
                        'status': 'low'
                    })
            elif isinstance(value, str):
                if len(value) > 100:
                    results['warnings'].append(f"Long string found in {key}")
                results['details'].append({
                    'type': 'string',
                    'length': len(value),
                    'key': key
                })
            elif isinstance(value, list):
                for i, item in enumerate(value):
                    if isinstance(item, dict):
                        for subkey, subvalue in item.items():
                            if isinstance(subvalue, (int, float)):
                                if subvalue < 0:
                                    results['errors'].append(f"Negative value in {key}[{i}].{subkey}")
                    elif isinstance(item, str) and len(item) > 50:
                        results['warnings'].append(f"Long string in list {key}[{i}]")
    
    # Apply options if provided
    if options:
        if 'threshold' in options:
            threshold = options['threshold']
            for key in list(results['summary'].keys()):
                if results['summary'][key] == 'high' and threshold == 'medium':
                    results['summary'][key] = 'medium'
                elif results['summary'][key] in ['high', 'medium'] and threshold == 'low':
                    results['summary'][key] = 'low'
        
        if 'filter' in options:
            filter_type = options['filter']
            if filter_type == 'warnings':
                results['details'] = [d for d in results['details'] if d.get('status') == 'warning']
            elif filter_type == 'errors':
                results['details'] = [d for d in results['details'] if d.get('status') == 'error']
    
    return results 